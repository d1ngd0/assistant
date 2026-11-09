package ui

import (
	"fmt"
	"net/http"

	"github.com/d1ngd0/assistant/state"
	"golang.org/x/net/websocket"
)

func SessionsHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		active := r.URL.Query().Get("active")
		websocket.Handler(HandleWSSessions(sessions, active)).ServeHTTP(w, r)
	})
}

func HandleWSSessions(sessions *state.Sessions, active string) websocket.Handler {
	return func(ws *websocket.Conn) {
		for {
			names := sessions.Names()
			SendWS(ws, "sessions.html", struct {
				Active string
				Names  []string
			}{
				Active: active,
				Names:  names,
			})

			select {
			case <-ws.Request().Context().Done():
				return
			case <-sessions.Wait():
			}
		}
	}

}

func DeleteSessionHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("session")
		_, ok := sessions.Get(name)
		if !ok {
			R404(w, r)
			return
		}

		sessions.Delete(name)
		w.WriteHeader(http.StatusNoContent)
	})
}

func AddSessionHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		name := r.FormValue("session")
		_, ok := sessions.Get(name)
		if ok {
			R400(w, fmt.Sprintf("%s already exists!", name))
			return
		}

		sessions.Fetch(name)
		w.WriteHeader(http.StatusNoContent)
	})
}
