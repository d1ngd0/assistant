package state

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/spf13/viper"
)

var (
	ErrModelNotFound = errors.New("model not found")
	ErrModelInvalid  = errors.New("model configuration invalid")
)

type ModelType string

type ModelRegistery struct {
	o      sync.Once
	cons   map[ModelType]ModelConstructor
	models map[string]Model
}

type ModelConstructor func(config *viper.Viper) (Model, error)

// NewModelRegistry returns a new model registry
func NewModelRegistry() *ModelRegistery {
	return &ModelRegistery{
		cons:   make(map[ModelType]ModelConstructor),
		models: make(map[string]Model),
	}
}

func (mr *ModelRegistery) init() {
	mr.o.Do(func() {
		mr.cons = make(map[ModelType]ModelConstructor)
		mr.models = make(map[string]Model)
	})
}

// Register registers a constructor with a specific model type. Now you
// can use this type whereever you want. congrats!
func (mr *ModelRegistery) Register(typ ModelType, cons ModelConstructor) {
	mr.init()
	mr.cons[typ] = cons
}

func (mr *ModelRegistery) Has(name string) bool {
	_, ok := mr.models[name]
	return ok
}

// Load is called to generate the model. This must be called in order for `Get` to return
// the model. There is no lazy loading.
func (mr *ModelRegistery) Load(name string, config *viper.Viper) error {
	mr.init()
	modelType := ModelType(config.GetString("type"))
	con, ok := mr.cons[modelType]
	if !ok {
		return fmt.Errorf("unknown type \"%s\": %w", modelType, ErrModelInvalid)
	}

	model, err := con(config)
	if err != nil {
		return err
	}

	mr.models[name] = model
	return nil
}

// Get returns the specified model. If the model doesn't exist it returns
// an error
func (mr *ModelRegistery) Get(name string) (Model, error) {
	mr.init()
	model, ok := mr.models[name]
	if !ok {
		return nil, fmt.Errorf("model \"%s\": %w", name, ErrModelNotFound)
	}

	return model, nil
}

type ResponseSync func(msg Message) error

// Model is the interface any implementation of model must
// implement to work with the assistent
type Model interface {
	// Send sends a message to the model. Many messages can be send with the last
	// being the most recent message to preserve historical context. Finally the
	// sync is a callback function used to write the response. Send can return as
	// soon as the request has begun.
	Send(ctx context.Context, msgs []Message, sync ResponseSync) error
}

type ModelFn func(ctx context.Context, msgs []Message, sync ResponseSync) error

func (mf ModelFn) Send(ctx context.Context, msgs []Message, sync ResponseSync) error {
	return mf(ctx, msgs, sync)
}
