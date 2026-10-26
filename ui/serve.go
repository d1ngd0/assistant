package ui

import (
	"log"
	"net/http"
	"slices"
)

// tmpl creates a set of templates from the
var (
	tmpl = must(loadTemplates())
	Log  *log.Logger
	ELog *log.Logger
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
func LoadRoutes(mux *http.ServeMux) *http.ServeMux {
	// Serve the data out of static
	mux.Handle("/static/", http.FileServerFS(Static))

	mux.Handle("/prompt", ExplicitPath(WithMethod(&Prompt{}, http.MethodPost), "/prompt"))
	mux.Handle("/session/{session}/context", http.HandlerFunc(ActiveSessions.ServeContext))

	// Set up the UI with the appropriate endpoints
	mux.Handle("/index", ExplicitPath(&Index{}, "/index"))
	mux.Handle("/", ExplicitPath(http.RedirectHandler("/index", http.StatusMovedPermanently), "/"))

	return mux
}

// explicitPath makes sure the path supplied is the one we are handling the request for
// if it isn't we return an error.
func ExplicitPath(handle http.Handler, path string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != path {
			R404(w, r)
			return
		}

		handle.ServeHTTP(w, r)
	})
}

// WithMethod runs the provided handler if the method is one of the provided
func WithMethod(handle http.Handler, methods ...string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if slices.Contains(methods, r.Method) {
			handle.ServeHTTP(w, r)
			return
		}

		R404(w, r)
	})
}

// R404 returns the 404 page
func R404(w http.ResponseWriter, r *http.Request) {
	// TODO: Make this pretty
	http.NotFound(w, r)
	_, _ = w.Write([]byte("oh fuck buddy"))
}

// R500 returns the 500 level error
func R500(w http.ResponseWriter, r *http.Request, err error) {
	http.Error(w, err.Error(), http.StatusInternalServerError)
}
