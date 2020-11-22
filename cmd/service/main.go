package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/url"
	"os"
	"os/signal"

	"github.com/WinPooh32/content/app"
	"github.com/WinPooh32/content/delivery"
	"github.com/WinPooh32/content/model"
	"github.com/WinPooh32/content/service"

	"github.com/go-chi/chi"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func osSignal() <-chan os.Signal {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	return c
}

func readTrackers(path string) ([]string, error) {
	var err error
	var file *os.File
	var trackers []string

	file, err = os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("os.Open: %w", err)
	}

	var scanner = bufio.NewScanner(file)
	scanner.Split(bufio.ScanLines)

	for scanner.Scan() {
		var line = scanner.Text()

		_, err = url.Parse(line)
		if err != nil {
			log.Warn().Err(err).Msgf("torrent url: %s", line)
			continue
		}

		log.Debug().Msgf("add tracker: %s", line)
		trackers = append(trackers, line)
	}

	return trackers, nil
}

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stdout})
}

func main() {
	var err error
	var r chi.Router

	var s = service.New()

	// Cmd flags.
	var host *string = flag.String("host", "127.0.0.1", "host")
	var port *uint = flag.Uint("port", 0, "port")
	var dir *string = flag.String("dir", "download", "root working directory")

	var trackersPath *string = flag.String("trackers", "trackers.txt", "path to trackers list file")
	var cacheCapacity *int64 = flag.Int64("size-cache", 10, "files cache capacity GiB")
	var readAhead *int64 = flag.Int64("size-readahead", 4, "readahead size MiB")
	var maxConn *int64 = flag.Int64("max-connections", 50, "max connection per torrent")
	var maxActive *int64 = flag.Int64("max-active", 4, "max active torrents")

	// Parse console arguments.
	flag.Parse()

	var settings = model.Settings{
		CacheSize:      *cacheCapacity << 30,
		ReadAheadSize:  *readAhead << 20,
		MaxConnections: *maxConn,
		MaxActive:      *maxActive,
	}

	var a *app.App
	var trackers []string

	trackers, err = readTrackers(*trackersPath)
	if err != nil {
		log.Warn().Msgf("failed to read trackers list: %s", err)
	}

	a, err = app.New(*dir, &settings, trackers)
	if err != nil {
		log.Fatal().Msgf("failed to create app: %s", err)
	}
	defer a.Close()

	r, err = delivery.NewHttpAPI(a)
	if err != nil {
		log.Fatal().Err(err).Msgf("init new http api")
	}

	err = s.Run(*host, uint16(*port), r)
	if err != nil {
		log.Fatal().Err(err).Msgf("http service run")
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
