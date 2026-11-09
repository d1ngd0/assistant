package state

type Message interface {
	Role() string
	Content() string
	Thought() string
}
