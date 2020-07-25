package igate

import (
	"errors"
	"net"
	"sync/atomic"
	"time"

	cot "github.com/nerdoftech/go-tak-proto/pkg/xml"
	aprs "github.com/pd0mz/go-aprs"
	"github.com/rs/zerolog/log"
)

type Server struct {
	udpConn     net.Conn
	tcpAddr     *net.TCPAddr
	tcpListener net.Listener
	ch          chan []byte
	stopped     int32
}

func NewServer(serverAddr string, udpHost string) (*Server, error) {
	errMsg := errors.New("could create new server")
	s := &Server{
		stopped: 1,
	}
	// CoT multicast UDP
	udpAddr, err := net.ResolveUDPAddr("udp", udpHost)
	if err != nil {
		log.Error().Err(err).Msg("failed to resolve UDP address")
		return nil, errMsg
	}

	s.udpConn, err = net.DialUDP("udp", nil, udpAddr)
	if err != nil {
		log.Error().Err(err).Msg("failed to make UDP conn")
		return nil, errMsg
	}

	// TCP igate server
	s.tcpAddr, err = net.ResolveTCPAddr("tcp", serverAddr)
	if err != nil {
		log.Error().Err(err).Msg("could not parse addr")
		return nil, errMsg
	}
	return s, nil
}

func (s *Server) Start() error {
	var err error
	s.tcpListener, err = net.ListenTCP("tcp", s.tcpAddr)
	if err != nil {
		msg := "could not start server"
		log.Error().Err(err).Msg(msg)
		return errors.New(msg)
	}
	s.stopped = 0

	s.ch = make(chan []byte, 10)
	go s.runServer()
	go s.handleAPRS()
	return nil
}

func (s *Server) Stop() {
	atomic.StoreInt32(&s.stopped, 1)
	s.tcpListener.Close()
	s.udpConn.Close()
}

func (s *Server) runServer() {
	ig := NewIGate(s.ch)
	for s.stopped == 0 {
		conn, err := s.tcpListener.Accept()
		if err != nil {
			log.Error().Err(err).Msg("incoming connection failed")
			continue
		}
		log.Debug().Interface("host", conn.RemoteAddr()).Msg("new client connection")
		go ig.HandleConn(conn)
	}
}

func (s *Server) handleAPRS() {
	for s.stopped == 0 {
		msg := <-s.ch
		pkt, err := aprs.ParsePacket(string(msg))
		if err != nil {
			log.Error().Err(err).Msg("could not parse packet")
			continue
		}
		log.Trace().Interface("pkt", pkt).Msg("new packet")

		if pkt.Position == nil {
			log.Error().Str("callsign", pkt.Src.String()).Msg("position not present in packet")
			log.Debug().Str("data", string(pkt.Payload)).Msg("payload")
			continue
		}

		log.Debug().
			Str("callsign", pkt.Src.String()).
			Str("type", pkt.Payload.Type().String()).
			Float64("lat", pkt.Position.Latitude).
			Float64("lon", pkt.Position.Longitude).
			Int("ambiguity", pkt.Position.Ambiguity).
			Float64("altitude", pkt.Altitude).
			Float64("course", pkt.Velocity.Course).
			Float64("speed", pkt.Velocity.Speed).
			Msg("packet contents")

		tk := &cot.Takv{
			OS:      "APRS",
			Version: "1",
			Device:  "radio",
		}
		cx := cot.NewCotXML(pkt.Src.String(), tk)
		cx.DfltStale = 10 * time.Minute

		pt := &cot.Point{
			Lat:  pkt.Position.Latitude,
			Long: pkt.Position.Longitude,
			Hae:  pkt.Altitude,
			CE:   3.0,
			LE:   9999.0,
		}
		tr := &cot.Track{
			Course: pkt.Velocity.Course,
			Speed:  pkt.Velocity.Speed,
		}
		loc := &cot.Loc{
			AltSrc: "APRS",
			Geo:    "APRS",
		}
		b, err := cx.OtherEvent(pkt.Src.String(), pkt.Src.String(), pt, tr, loc).MarshallEvent()
		if err != nil {
			log.Error().Err(err).Msg("failed to marshall CoT")
			continue
		}

		sz, err := s.udpConn.Write(b)
		if err != nil {
			log.Error().Err(err).Msg("could not send data")
		}
		log.Debug().Int("size", sz).Msg("bytes written to socket")
	}
}
