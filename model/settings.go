package model

import (
	"fmt"
	"net/http"
)

type Settings struct {
	MaxConnections int64 `json:"max_connections"`
	CacheSize      int64 `json:"cache_size"`
	ReadAheadSize  int64 `json:"read_ahead"`
	MaxActive      int64 `json:"max_active"`
}

func (s *Settings) Render(w http.ResponseWriter, r *http.Request) error {
	return nil
}

func (s *Settings) Bind(r *http.Request) error {

	if s.MaxConnections < 0 {
		return fmt.Errorf("settings: max_connections: value out of bounds: %d", s.MaxConnections)
	}

	if s.CacheSize < 0 {
		return fmt.Errorf("settings: cache_size: out of bounds: %d", s.CacheSize)
	}

	if s.ReadAheadSize < 0 {
		return fmt.Errorf("settings: read_ahead: out of bounds: %d", s.ReadAheadSize)
	}

	if s.MaxActive < 0 {
		return fmt.Errorf("settings: max_active: out of bounds: %d", s.MaxActive)
	}

	return nil
}
