package api

import (
	"encoding/json"
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

type Message struct {
	Content string `json:"content"`
	Role    string `json:"role"`
	Thought string `json:"thought"`
}

func ContextHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		name := req.PathValue("session")
		if name == "" {
			R404(rw)
			return
		}

		session, ok := sessions.Get(name)
		if !ok {
			R404(rw)
			return
		}

		var context []Message
		for msg := range session.Contexts(0) {
			context = append(context, Message{
				Content: msg.Content(),
				Role:    msg.Role(),
				Thought: msg.Thought(),
			})
		}

		b, err := json.Marshal(context)
		if err != nil {
			R500(rw, err)
			return
		}

		_, _ = rw.Write(b)
	})
}
