package main

import (
	"content/app"
	"content/delivery"
	"content/service"
	"flag"
	"os"
	"os/signal"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func osSignal() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	return c
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func main() {
	var err error
	var r chi.Router

	var s = service.New()

	var host *string = flag.String("host", "127.0.0.1", "host")
	var port *uint = flag.Uint("port", 0, "port")

	// Parse console arguments.
	flag.Parse()

	var a *app.App

	a, err = app.New(nil)
	if err != nil {
		log.Fatal().Msgf("failed to create app: %s", err)
	}
	defer a.Close()

	r, err = delivery.NewHttpAPI(a)
	if err != nil {
		log.Error().Err(err).Msgf("init new http api")
		return
	}

	err = s.Run(*host, uint16(*port), r)
	if err != nil {
		log.Error().Err(err).Msgf("http service run")
		return
	}

	select {
	case <-osSignal():
		s.Stop()
	case err = <-s.Done():
		if err != nil {
			log.Error().Err(err).Msgf("service error")
		}
	}

	log.Info().Msg("exit")
}
