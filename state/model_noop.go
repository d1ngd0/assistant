package state

import (
	"context"
)

const ModelTypeNoOp = "noop"

const RoleUser = "user"

type Msg struct {
	role    string
	content string
	thought string
}

func StringMsg(msg string) *Msg {
	return &Msg{
		role:    RoleUser,
		content: msg,
		thought: "",
	}
}

func NewMsg(role, content, thought string) *Msg {
	return &Msg{
		role:    role,
		content: content,
		thought: thought,
	}
}

// Content implements [Message].
func (s *Msg) Content() string {
	return s.content
}

// Role implements [Message].
func (s *Msg) Role() string {
	return s.role
}

// Thought implements [Message].
func (s *Msg) Thought() string {
	return s.thought
}

var _ Message = (*Msg)(nil)

func NoOpModel(ctx context.Context, msgs []Message, sync ResponseSync) error {
	return sync(StringMsg("This is a no op model"))
}
