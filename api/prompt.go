package api

import (
	"encoding/json"
	"github.com/d1ngd0/assistant/state"
	"io"
	"net/http"
)

// Prompt is the payload the prompt handler takes.
type Prompt struct {
	Session string `json:"session"`
	Model   string `json:"model"`
	Message string `json:"message"`
}

// PromptHandler takes a [Prompt] payload as a json body
// URI: /prompt
// Methods: PUT, POST
func PromptHandler(sessions *state.Sessions, models *state.ModelRegistery) http.Handler {
	return WithMethod(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {

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

		session := sessions.Fetch(prompt.Session)
		model, err := models.Get(prompt.Model)
		if err != nil {
			R500(w, err)
			return
		}

		err = model.Send(req.Context(), session.Push(state.Msg(prompt.Message)), session.Sync)
		if err != nil {
			R500(w, err)
			return
		}
	}), http.MethodPut, http.MethodPost)
}
