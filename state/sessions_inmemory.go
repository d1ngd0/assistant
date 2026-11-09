package state

import (
	"iter"
	"slices"
	"sync"
)

type Sessions struct {
	o        sync.Once
	lock     sync.RWMutex
	event    chan struct{}
	Sessions map[string]*Session
}

func (s *Sessions) init() {
	s.o.Do(func() {
		s.Sessions = make(map[string]*Session)
		s.event = make(chan struct{})
	})
}

// Names returns the names in the current session
func (s *Sessions) Names() []string {
	var names []string
	for key := range s.Sessions {
		names = append(names, key)
	}

	slices.Sort(names)

	return names
}

// Get returns the session
func (s *Sessions) Get(name string) (*Session, bool) {
	s.init()

	s.lock.RLock()
	session, ok := s.Sessions[name]
	s.lock.RUnlock()
	return session, ok
}

func (s *Sessions) Fetch(name string) *Session {
	session, ok := s.Get(name)

	if !ok {
		s.lock.Lock()
		// check again, someone might have created
		// it while we were reading
		session, ok = s.Sessions[name]
		if !ok {
			session = newSession()
			s.Sessions[name] = session
			close(s.event)
			s.event = make(chan struct{})
		}
		s.lock.Unlock()
	}

	return session
}

func (s *Sessions) Delete(name string) {
	s.init()

	s.lock.Lock()
	defer s.lock.Unlock()

	delete(s.Sessions, name)
	close(s.event)
	s.event = make(chan struct{})
}

// Wait blocks until a new session has appeared.
func (s *Sessions) Wait() chan struct{} {
	s.lock.RLock()
	c := s.event
	s.lock.RUnlock()
	return c
}

type Session struct {
	event   chan struct{}
	lock    sync.RWMutex
	context []Message
}

func newSession() *Session {
	s := &Session{
		event: make(chan struct{}),
	}
	return s
}

// Push places the latest message into the context
func (s *Session) Push(msg Message) []Message {
	s.lock.Lock()
	defer s.lock.Unlock()

	close(s.event)
	s.event = make(chan struct{})

	s.context = append(s.context, msg)
	return s.context
}

// Sync is a [ResponseSync] implementation that is used to store the response
// in the context.
func (s *Session) Sync(msg Message) error {
	_ = s.Push(msg)
	return nil
}

// Contexts returns an iterator which locks the session context until it is exhausted
func (s *Session) Contexts(from int) iter.Seq2[Message, int] {
	s.lock.RLock()
	return func(yield func(msg Message, index int) bool) {
		defer s.lock.RUnlock()

		for x := from; x < len(s.context); x++ {
			if !yield(s.context[x], x) {
				return
			}
		}
	}
}

// Clear removes all the context for the session
func (s *Session) Reset(msg ...Message) {
	s.lock.Lock()
	defer s.lock.Unlock()

	close(s.event)
	s.event = make(chan struct{})
	s.context = msg
}

// return a channel which will return when the next event has occured. This calls
// an internal lock to fetch the appropriate channel.
func (s *Session) Wait() chan struct{} {
	s.lock.RLock()
	c := s.event
	s.lock.RUnlock()
	return c
}

/// This next section stores all the code for maintainning the websockets
/// currently connected to the session
