package api

import (
	"encoding/json"
	"io"
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

type SessionPayload struct {
	Name    string    `json:"name"`
	Context []Message `json:"context"`
}

func (sp *SessionPayload) Validate() error {
	if sp.Name == "" {
		return requiredError("name")
	}

	return nil
}

func ExportSessionHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		var payload SessionPayload
		payload.Name = req.PathValue("session")
		if payload.Name == "" {
			R404(rw)
			return
		}

		session, exists := sessions.Get(payload.Name)
		if !exists {
			R404(rw)
			return
		}

		for msg := range session.Contexts(0) {
			payload.Context = append(payload.Context, Message{
				Content: msg.Content(),
				Role:    msg.Role(),
				Thought: msg.Thought(),
			})
		}

		b, err := json.Marshal(payload)
		if err != nil {
			R500(rw, err)
			return
		}

		_, _ = rw.Write(b)
	})
}

func ImportSessionHandler(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			R500(rw, err)
			return
		}

		var payload SessionPayload
		if err := json.Unmarshal(b, &payload); err != nil {
			R500(rw, err)
			return
		}

		session := sessions.Fetch(payload.Name)
		var context []state.Message
		for _, msg := range payload.Context {
			context = append(context, state.NewMsg(msg.Role, msg.Content, msg.Thought))
		}
		session.Reset(context...)
		rw.WriteHeader(http.StatusNoContent)
	})
}
