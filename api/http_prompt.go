package api

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/d1ngd0/assistant/state"
)

// Prompt is the payload the prompt handler takes.
type Prompt struct {
	Session string `json:"session"`
	Model   string `json:"model"`
	Message string `json:"message"`
}

func (p *Prompt) Validate() error {
	if p.Session == "" {
		return requiredError("session")
	}

	if p.Model == "" {
		return requiredError("model")
	}

	if p.Message == "" {
		return requiredError("message")
	}

	return nil
}

// PromptHandler takes a [Prompt] payload as a json body
// URI: /prompt
// Methods: PUT, POST
func PromptHandler(sessions *state.Sessions, models *state.ModelRegistery) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		b, err := io.ReadAll(req.Body)
		if err != nil {
			R500(w, err)
			return
		}

		var prompt Prompt
		if err := json.Unmarshal(b, &prompt); err != nil {
			R500(w, err)
			return
		}

		if err := prompt.Validate(); err != nil {
			R400(w, err.Error())
			return
		}

		session := sessions.Fetch(prompt.Session)
		model, err := models.Get(prompt.Model)
		if err != nil {
			switch {
			case errors.Is(err, state.ErrModelNotFound):
				R400(w, err.Error())
			default:
				R500(w, err)
			}
			return
		}

		err = model.Send(req.Context(), session.Push(state.Msg(prompt.Message)), session.Sync)
		if err != nil {
			R500(w, err)
			return
		}
	})
}
