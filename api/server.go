package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

// tmpl creates a set of templates from the
var (
	Log  *log.Logger
	ELog *log.Logger

	ErrRequired = errors.New("required field not specified")
)

// must returns the value when no error is present, otherwise it panics
func must[T any](v T, err error) T {
	if err != nil {
		panic(err.Error())
	}
	return v
}

// LoadRoutes specifies all of the route available for the ui. If the ui has a page
// and you want to find it, it better be in this list. Nothing is hidden from you.
func LoadRoutes(mux *http.ServeMux, models *state.ModelRegistery, sessions *state.Sessions) *http.ServeMux {
	mux.Handle("/session/{session}/context", MapMethod(
		WithMethod(http.MethodGet, ContextHandler(sessions)),
		WithMethod(http.MethodPost, ContextAppendHandler(sessions)),
	))
	mux.Handle("/sessions", MapMethod(
		WithMethod(http.MethodGet, SessionsHandler(sessions)),
		WithMethod(http.MethodPost, ImportSessionHandler(sessions)),
	))

	mux.Handle("/session/{session}", MapMethod(
		WithMethod(http.MethodGet, ExportSessionHandler(sessions)),
	))

	mux.Handle("/prompt", MapMethod(
		WithMethod(http.MethodPost, PromptHandler(sessions, models)),
	))

	mux.Handle("/", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		R404(w)
	}))

	return mux
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func ErrString(msg string) ErrorResponse {
	return ErrorResponse{
		Error: msg,
	}
}

func errResponse(rw http.ResponseWriter, err ErrorResponse, code int) {
	rw.WriteHeader(code)
	b := must(json.Marshal(err))
	_, _ = rw.Write(b)
}

// R404 is the 404 response for the API
func R404(rw http.ResponseWriter) {
	errResponse(rw, ErrString("not found"), http.StatusNotFound)
}

// R404 is the 404 response for the API
func R500(rw http.ResponseWriter, err error) {
	errResponse(rw, ErrString(err.Error()), http.StatusInternalServerError)
}

func R400(rw http.ResponseWriter, err error) {
	errResponse(rw, ErrString(err.Error()), http.StatusBadRequest)
}

type MethodBundle struct {
	method  string
	handler http.Handler
}

func WithMethod(method string, handler http.Handler) MethodBundle {
	return MethodBundle{
		method:  method,
		handler: handler,
	}
}

func MapMethod(handlers ...MethodBundle) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for _, handler := range handlers {
			if r.Method == handler.method {
				handler.handler.ServeHTTP(w, r)
				return
			}
		}

		R404(w)
	})
}

func requiredError(field string) error {
	return fmt.Errorf("%w: %s", ErrRequired, field)
}
