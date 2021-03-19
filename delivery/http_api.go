package delivery

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/WinPooh32/content/app"
	"github.com/WinPooh32/content/model"

	"github.com/anacrolix/torrent"
	"github.com/asaskevich/govalidator"
	"github.com/go-chi/chi"
	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type urlParam string

const (
	paramHash urlParam = "hash"
	paramPath urlParam = "path"
)

type API struct {
	app *app.App
}

func NewHttpAPI(app *app.App) (chi.Router, error) {

	var api = API{
		app: app,
	}

	var r = chi.NewRouter()

	r.Get("/ping", api.pingGET)

	r.Route("/settings", func(r chi.Router) {
		r.Get("/", api.senttingsGET)
		r.Put("/", api.senttingsPUT)
	})

	r.With(hash).Route(fmt.Sprintf("/content/{%s}", paramHash), func(r chi.Router) {
		r.With(path).Get("/*", api.contentGET)
		r.Get("/", api.contentPUT)
		r.Get("/info", api.contentInfoGET)
	})

	return r, nil
}

func (api *API) pingGET(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintln(w, "OK")
}

func (api *API) senttingsGET(w http.ResponseWriter, r *http.Request) {

}

func (api *API) senttingsPUT(w http.ResponseWriter, r *http.Request) {

}

func (api *API) contentGET(w http.ResponseWriter, r *http.Request) {
	var err error
	var t *torrent.Torrent

	var hex string = r.Context().Value(paramHash).(string)
	var path string = r.Context().Value(paramPath).(string)

	log.Info().Msgf("GET %s/%s", hex, path)

	t, err = api.app.TrackHash(r.Context(), hex)
	if err != nil {
		_ = render.Render(w, r, ErrInternal(fmt.Errorf("failed to track torrent: %w", err)))
		return
	}

	select {
	case <-r.Context().Done():
		http.Error(w, http.StatusText(http.StatusRequestTimeout), http.StatusRequestTimeout)
		return

	case <-t.GotInfo():
	}

	err = api.serveTorrentFile(w, r, t, path)
	if err != nil {
		_ = render.Render(w, r, ErrInternal(fmt.Errorf("failed to serve torrent file: %w", err)))
		return
	}
}

func (api *API) contentPUT(w http.ResponseWriter, r *http.Request) {
	var err error
	var t *torrent.Torrent

	var hex = r.Context().Value(paramHash).(string)

	t, err = api.app.TrackHash(r.Context(), hex)
	if err != nil {
		_ = render.Render(w, r, ErrInternal(fmt.Errorf("failed to track torrent: %w", err)))
		return
	}

	err = render.Render(w, r, model.NewFilesList(t))
	if err != nil {
		_ = render.Render(w, r, ErrInternal(fmt.Errorf("failed to make files list from torrent: %w", err)))
		return
	}
}

func (api *API) contentInfoGET(w http.ResponseWriter, r *http.Request) {
	var err error
	var t *torrent.Torrent
	var ok bool

	var hex = r.Context().Value(paramHash).(string)

	t, ok = api.app.Torrent(hex)
	if !ok {
		_ = render.Render(w, r, ErrNotFound(fmt.Errorf("failed to find torrent: %s", hex)))
		return
	}

	err = render.Render(w, r, model.NewFilesList(t))
	if err != nil {
		_ = render.Render(w, r, ErrInternal(fmt.Errorf("failed to make files list from torrent: %w", err)))
		return
	}
}

func hash(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var hash = strings.ToLower(chi.URLParam(r, "hash"))

		if !govalidator.IsSHA1(hash) {
			if err := render.Render(w, r, ErrBadRequest(fmt.Errorf("incorrect hash value"))); err != nil {
				log.Error().Err(err).Msgf("chi render")
			}
			return
		}

		ctx := context.WithValue(r.Context(), paramHash, hash)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func path(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var path, err = url.QueryUnescape(chi.URLParam(r, "*"))
		if err != nil {
			if err = render.Render(w, r, ErrBadRequest(fmt.Errorf("url unescape"))); err != nil {
				log.Error().Err(err).Msgf("chi render")
				return
			}
		}

		if !govalidator.IsRequestURI("/" + path) {
			if err := render.Render(w, r, ErrBadRequest(fmt.Errorf("invalid path"))); err != nil {
				log.Error().Err(err).Msgf("chi render")
			}
			return
		}

		ctx := context.WithValue(r.Context(), paramPath, path)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}
