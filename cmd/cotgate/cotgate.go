package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/nerdoftech/go-igate-cot/igate"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	DEFAULT_PORT = 14580
	udpMulticast = "239.2.3.1:6969"
)

var (
	flgAddr   = flag.String("addr", "0.0.0.0", "The local IP address for the bootnode discovery server")
	flgPort   = flag.Int("port", DEFAULT_PORT, "UDP port for Cot IGate server")
	flgLogLvl = flag.String("log", "info", "Sets the log level")
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	flag.Parse()
	// Set log level
	lvl, err := zerolog.ParseLevel(*flgLogLvl)
	if err != nil {
		log.Fatal().Err(err).Msg("could not parse log level")
	}
	zerolog.SetGlobalLevel(lvl)

	igateAddr := fmt.Sprintf("%s:%d", *flgAddr, *flgPort)
	svr, err := igate.NewServer(igateAddr, udpMulticast)
	if err != nil {
		log.Fatal().Msg(err.Error())
	}
	svr.Start()

	log.Debug().Msgf("server started, waiting for signal to shutdown")
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	sig := <-sigCh
	log.Info().Interface("signal", sig).Msg("got signal, shutting down")
	svr.Stop()
}
