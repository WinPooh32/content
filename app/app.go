package app

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"content/model"

	anacrolixlog "github.com/anacrolix/log"
	"github.com/anacrolix/sync"
	"github.com/anacrolix/torrent"
	"github.com/anacrolix/torrent/metainfo"
	"github.com/anacrolix/torrent/mse"
	"github.com/rs/zerolog/log"

	"github.com/boltdb/bolt"
)

const (
	dbName       = ".app.db"
	dbBucketInfo = "info"
)

type App struct {
	client *torrent.Client
	sets   *model.Settings
	db     *bolt.DB

	torrents map[string]*torrent.Torrent
	mu       sync.RWMutex

	trackers [][]string
}

func (app *App) TrackMagnet(ctx context.Context, magnet *metainfo.Magnet) (*torrent.Torrent, error) {
	var err error
	var t *torrent.Torrent

	var spec = &torrent.TorrentSpec{
		Trackers:    append(app.trackers, magnet.Trackers),
		DisplayName: magnet.DisplayName,
		InfoHash:    magnet.InfoHash,
	}

	t, _, err = app.client.AddTorrentSpec(spec)
	if err != nil {
		return nil, fmt.Errorf("torrent add magnet: %w", err)
	}

	return t, app.trackContext(ctx, t)
}

func (app *App) Torrent(hash string) (*torrent.Torrent, bool) {
	app.mu.RLock()
	defer app.mu.RUnlock()

	t, ok := app.torrents[hash]
	return t, ok
}

func (app *App) Close() error {
	var err error

	// Close database.
	if app.db != nil {
		err = app.db.Close()
		if err != nil {
			return fmt.Errorf("close db: %w", err)
		}
	}

	return nil
}

func (app *App) trackContext(ctx context.Context, t *torrent.Torrent) error {

	select {
	case <-ctx.Done():
		return ctx.Err()

	case <-t.GotInfo():
	}

	var err = app.track(t)
	if err != nil {
		return fmt.Errorf("track torrent: %w", err)
	}

	return nil
}

func (app *App) track(t *torrent.Torrent) error {
	var err error

	err = app.db.Update(func(tx *bolt.Tx) error {
		var err error
		var mi = t.Metainfo()
		var buf = bytes.NewBuffer(nil)

		err = mi.Write(buf)
		if err != nil {
			return fmt.Errorf("write metaInfo: %w", err)
		}

		var b = tx.Bucket([]byte(dbBucketInfo))
		var ih = (t.InfoHash())

		return b.Put(ih.Bytes(), buf.Bytes())
	})
	if err != nil {
		return fmt.Errorf("put to db: %w", err)
	}

	app.mu.Lock()
	app.torrents[t.InfoHash().String()] = t
	app.mu.Unlock()

	return nil
}

func (app *App) untrack(t *torrent.Torrent) error {
	var err error

	err = app.db.Update(func(tx *bolt.Tx) error {
		var b = tx.Bucket([]byte(dbBucketInfo))
		return b.Delete(t.InfoHash().Bytes())
	})
	if err != nil {
		return fmt.Errorf("put to db: %w", err)
	}

	app.mu.Lock()
	delete(app.torrents, t.InfoHash().String())
	app.mu.Unlock()

	t.Drop()

	return nil
}

func (app *App) load() error {
	app.mu.Lock()
	defer app.mu.Unlock()

	return app.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketInfo))

		return b.ForEach(func(k, v []byte) error {
			var err error
			var mi *metainfo.MetaInfo
			var t *torrent.Torrent

			mi, err = metainfo.Load(bytes.NewReader(v))
			if err != nil {
				log.Warn().Msgf("read meta info: %s", err)
				return nil
			}

			t, err = app.client.AddTorrent(mi)
			if err != nil {
				log.Warn().Msgf("add torrent: %s", err)
				return nil
			}

			app.torrents[t.InfoHash().String()] = t

			return nil
		})
	})
}

func newTorrentSettings(sets *model.Settings) *torrent.ClientConfig {
	var cfg *torrent.ClientConfig = torrent.NewDefaultClientConfig()

	// Take random free port.
	cfg.ListenPort = 0

	// Enable seeding.
	cfg.Seed = true

	// Header obfuscation.
	cfg.HeaderObfuscationPolicy = torrent.HeaderObfuscationPolicy{
		Preferred:        true,
		RequirePreferred: true,
	}

	// Force encryption.
	cfg.CryptoProvides = mse.CryptoMethodRC4

	cfg.DefaultRequestStrategy = torrent.RequestStrategyFastest()

	// Torrent debug.
	cfg.Debug = false
	cfg.Logger = anacrolixlog.Discard

	return cfg
}

func openDB(path string) (*bolt.DB, error) {
	var db, err = bolt.Open(path, 0600, &bolt.Options{
		Timeout: 30 * time.Second,
	})
	if err != nil {
		return nil, err
	}

	// Create buckets.
	err = db.Update(func(tx *bolt.Tx) error {
		var _, err = tx.CreateBucketIfNotExists([]byte(dbBucketInfo))
		if err != nil {
			return fmt.Errorf("create bucket: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	return db, nil
}

func New(sets *model.Settings, trackers []string) (*App, error) {
	var err error
	var t *torrent.Client
	var defaultSets model.Settings
	var store *bolt.DB

	if sets != nil {
		defaultSets = *sets
	}

	t, err = torrent.NewClient(newTorrentSettings(&defaultSets))
	if err != nil {
		return nil, fmt.Errorf("failed to create new torrent client: %w", err)
	}

	store, err = openDB(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	var app = &App{
		client:   t,
		sets:     &defaultSets,
		db:       store,
		torrents: make(map[string]*torrent.Torrent),
		trackers: [][]string{{"http://retracker.local/announce"}, trackers},
	}

	return app, app.load()
}
