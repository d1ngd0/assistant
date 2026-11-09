package ui

import (
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

func ServePrompt(models *state.ModelRegistery, sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		input := req.FormValue("prompt")
		Log.Printf("recieved prompt %s", input)

		session := sessions.Fetch(req.FormValue("session"))
		model, err := models.Get(req.FormValue("model"))
		if err != nil {
			R500(w, req, err)
			return
		}

		err = model.Send(req.Context(), session.Push(state.StringMsg(input)), session.Sync)
		if err != nil {
			R500(w, req, err)
			return
		}
	})
}
