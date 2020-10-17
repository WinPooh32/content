package service

import (
	"fmt"
	"net"
	"net/http"
	"sync"

	"github.com/rs/zerolog/log"
)

type Service struct {
	http   http.Server
	done   chan error
	finish sync.Once
}

// run - runs http server.
func (s *Service) Run(host string, port uint16, handler http.Handler) error {

	s.http = http.Server{
		Handler: handler,
	}

	var addr = fmt.Sprintf("%s:%d", host, port)
	var ln, err = net.Listen("tcp", addr)
	if err != nil {
		return fmt.Errorf("net.Listen tcp %s: %w", addr, err)
	}

	log.Info().Msgf("http listen at http://%s", ln.Addr().String())

	go func() {
		var err = s.http.Serve(ln)
		if err != nil && err == http.ErrServerClosed {
			s.done <- fmt.Errorf("http listen and serve: %w", err)
		}
	}()

	return err
}

func (s *Service) Stop() {
	s.finish.Do(func() {
		s.done <- nil
	})
}

func (s *Service) Done() <-chan error {
	return s.done
}

func New() *Service {
	return &Service{
		done: make(chan error, 1),
	}
}
