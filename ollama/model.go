package ollama

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/d1ngd0/assistant/state"
	"github.com/ollama/ollama/api"
)

const ModelType = "ollama"

type Config struct {
	Url   string `json:"url"`
	Model string `json:"model"`
	Think bool   `json:"think"`
}

type Model struct {
	client *api.Client
	model  string
	think  bool
}

// NewOllamaModel returns a new model with the following url and http client
func NewOllamaModel(client *http.Client, config Config) (*Model, error) {
	if client == nil {
		client = http.DefaultClient
	}

	u, err := url.Parse(config.Url)
	if err != nil {
		return nil, fmt.Errorf(`url "%s"`, config.Url)
	}

	return &Model{
		client: api.NewClient(u, client),
		model:  config.Model,
		think:  config.Think,
	}, nil
}

// Send implements [state.Model].
func (o *Model) Send(ctx context.Context, msgs []state.Message, sync state.ResponseSync) error {
	var messages []api.Message
	for x := range msgs {
		messages = append(messages, api.Message{
			Role:     msgs[x].Role(),
			Content:  msgs[x].Content(),
			Thinking: msgs[x].Thought(),
		})
	}

	return o.client.Chat(ctx, &api.ChatRequest{
		Model:    o.model,
		Messages: messages,
		Stream:   new(false),
		Think:    &api.ThinkValue{Value: o.think},
	}, func(resp api.ChatResponse) error {
		return sync(Message{resp.Message})
	})
}

type Message struct {
	child api.Message
}

// Content implements [ui.Message].
func (m Message) Content() string {
	return m.child.Content
}

// Role implements [ui.Message].
func (m Message) Role() string {
	return m.child.Role
}

// Thought implements [ui.Message].
func (m Message) Thought() string {
	return m.child.Thinking
}

var _ state.Model = (*Model)(nil)
var _ state.Message = (*Message)(nil)
