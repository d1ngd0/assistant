package ollama

import (
	"context"
	"net/http"

	"github.com/d1ngd0/assistant/ui"
)

type OllamaModel struct {
	client   http.Client
	endpoint string
	model    string
}

// Send implements [ui.Model].
func (o *OllamaModel) Send(ctx context.Context, msgs []string, sync ui.ResponseSync) error {
	panic("unimplemented")
}

var _ ui.Model = (*OllamaModel)(nil)
