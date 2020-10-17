package delivery

import (
	"net/http"

	"github.com/go-chi/render"
	"github.com/rs/zerolog/log"
)

type errStatus string

const (
	errStatusNotFound      errStatus = "not found"
	errStatusBadRequest    errStatus = "bad request"
	errStatusInternal      errStatus = "internal error"
	errStatusUnprocessable errStatus = "failed to proccess request"
)

type ErrResponse struct {
	Err            error `json:"-"` // low-level runtime error
	HTTPStatusCode int   `json:"-"` // http response status code

	StatusText errStatus `json:"status"`          // user-level status message
	AppCode    int64     `json:"code,omitempty"`  // application-specific error code
	ErrorText  string    `json:"error,omitempty"` // application-level error message, for debugging
}

func (rsp *ErrResponse) Render(w http.ResponseWriter, r *http.Request) error {
	// Pre-processing before a response is marshalled and sent across the wire
	render.Status(r, rsp.HTTPStatusCode)
	return nil
}

func ErrNotFound(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusNotFound,
		StatusText:     errStatusNotFound,
		ErrorText:      err.Error(),
	}
}

func ErrBadRequest(err error) render.Renderer {
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusBadRequest,
		StatusText:     errStatusBadRequest,
		ErrorText:      err.Error(),
	}
}

func ErrInternal(err error) render.Renderer {
	log.Error().Err(err).Msgf("runtime internal error")
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusInternalServerError,
		StatusText:     errStatusInternal,
		ErrorText:      err.Error(),
	}
}

func ErrRender(err error) render.Renderer {
	log.Error().Err(err).Msgf("failed to render response")
	return &ErrResponse{
		Err:            err,
		HTTPStatusCode: http.StatusUnprocessableEntity,
		StatusText:     errStatusUnprocessable,
		ErrorText:      err.Error(),
	}
}
