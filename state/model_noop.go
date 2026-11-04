package state

import (
	"context"
)

const ModelTypeNoOp = "noop"

func init() {
}

type StringMessage struct {
	msg string
}

func Msg(msg string) StringMessage {
	return StringMessage{msg: msg}
}

// Content implements [Message].
func (s StringMessage) Content() string {
	return s.msg
}

// Role implements [Message].
func (s StringMessage) Role() string {
	return "user"
}

// Thought implements [Message].
func (s StringMessage) Thought() string {
	return ""
}

var _ Message = StringMessage{msg: ""}

func NoOpModel(ctx context.Context, msgs []Message, sync ResponseSync) error {
	return sync(StringMessage{msg: "This is a no op model"})
}
