package api

import (
	"encoding/json"
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

func SessionsHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		names := sessions.Names()
		if names == nil {
			names = make([]string, 0)
		}

		b, err := json.Marshal(names)
		if err != nil {
			R500(w, err)
			return
		}

		// TODO: Add pagination logic

		_, _ = w.Write(b)
	})
}
