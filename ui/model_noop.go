package ui

import (
	"context"

	"github.com/spf13/viper"
)

const ModelTypeNoOp = "noop"

func init() {
	Models.Register(ModelTypeNoOp, func(config *viper.Viper) (Model, error) {
		return ModelFn(NoOpModel), nil
	})
}

func NoOpModel(ctx context.Context, msgs []string, sync ResponseSync) error {
	return sync("This is a no op model")
}
