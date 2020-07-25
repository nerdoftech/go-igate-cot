package igate

import (
	"net"
	"strings"
	"sync"

	"github.com/rs/zerolog/log"
)

const (
	HB       byte = 0x23
	NL       byte = 0x0a
	RTN      byte = 0x0d
	BUF_SIZE      = 512
)

type buff struct {
	buf     []byte
	cnt     int
	rtn     bool
	readIdx int
}

func (b *buff) reset() {
	b.buf = make([]byte, BUF_SIZE)
	b.cnt = 0
	b.readIdx = 0
	b.rtn = false
}

func (b *buff) readBytes() []byte {
	sl := b.buf[b.readIdx:b.cnt]
	b.readIdx = b.cnt
	return sl
}

func (b *buff) writeByte(char byte) {
	b.buf[b.cnt] = char
	b.cnt++
}

// IGate
type IGate struct {
	msgChan chan []byte
	mu      sync.Mutex
}

// NewIGate returns IGate
func NewIGate(ch chan []byte) *IGate {
	return &IGate{
		msgChan: ch,
	}
}

// HandleConn .
func (ig *IGate) HandleConn(conn net.Conn) {
	authDone := false
	bf := &buff{}
	bf.reset()
	for {
		b := make([]byte, 1)
		_, err := conn.Read(b)
		if err != nil {
			log.Debug().Err(err).Interface("host", conn.RemoteAddr()).Msg("could not read bytes")
			conn.Close()
			return
		}

		switch b[0] {
		case RTN: // \r\n means end of message
			bf.rtn = true
			continue
		case NL:
			// First packet should be auth
			if !authDone {
				if handleAuth(bf.readBytes()) {
					authDone = true
					bf.reset()
					continue
				} else { // Close conn if auth fails
					log.Error().
						Interface("host", conn.RemoteAddr()).
						Msg("login failed")
					conn.Close()
					return
				}
			}
			if bf.rtn {
				// FIXME - handle empty buf count

				// Did we get a heartbeat?
				if bf.cnt == 1 && bf.buf[0] == HB {
					log.Trace().
						Interface("host", conn.RemoteAddr()).
						Msg("heartbeat")
					bf.reset()
					continue
				}

				ig.mu.Lock()
				ig.msgChan <- bf.readBytes()
				ig.mu.Unlock()
				bf.reset()
				continue
			}

		}
		bf.writeByte(b[0])
		if bf.cnt >= BUF_SIZE {
			log.Debug().Int("len", BUF_SIZE).Interface("host", conn.RemoteAddr()).Msg("message too big")
			bf.reset()
		}
	}
}

// TODO, add login storage backend
func handleAuth(b []byte) bool {
	login := strings.Split(string(b), " ")
	// Validate login packet
	if len(login) < 4 {
		return false
	}
	if login[0] != "user" {
		return false
	}
	if login[2] != "pass" {
		return false
	}
	log.Info().Str("user", login[1]).Msg("new login")
	return true
}
