package ui

import (
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

type Prompt struct {
	models   *state.ModelRegistery
	sessions *Sessions
}

func (p *Prompt) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	input := req.FormValue("prompt")
	Log.Printf("recieved prompt %s", input)

	session := p.sessions.Fetch(req.FormValue("session"))
	model, err := p.models.Get(req.FormValue("model"))
	if err != nil {
		R500(w, req, err)
		return
	}

	err = model.Send(req.Context(), session.Push(state.Msg(input)), session.Sync)
	if err != nil {
		R500(w, req, err)
		return
	}
}
