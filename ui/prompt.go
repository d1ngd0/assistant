package ui

import (
	"net/http"
)

type Prompt struct{}

func (p *Prompt) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	input := req.FormValue("prompt")
	Log.Printf("recieved prompt %s", input)

	session := ActiveSessions.Fetch(req.FormValue("session"))
	model, err := Models.Get(req.FormValue("model"))
	if err != nil {
		R500(w, req, err)
		return
	}

	err = model.Send(req.Context(), session.Push(input), session.Sync)
	if err != nil {
		R500(w, req, err)
		return
	}
}
