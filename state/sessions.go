package state

import (
	"iter"
	"slices"
	"sync"
)

type Sessions struct {
	o        sync.Once
	Lock     sync.RWMutex
	Sessions map[string]*Session
}

func (s *Sessions) init() {
	s.o.Do(func() {
		s.Sessions = make(map[string]*Session)
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

	s.Lock.RLock()
	session, ok := s.Sessions[name]
	s.Lock.RUnlock()
	return session, ok
}

func (s *Sessions) Fetch(name string) *Session {
	session, ok := s.Get(name)

	if !ok {
		s.Lock.Lock()
		// check again, someone might have created
		// it while we were reading
		session, ok = s.Sessions[name]
		if !ok {
			session = newSession()
			s.Sessions[name] = session
		}
		s.Lock.Unlock()
	}

	return session
}

func (s *Sessions) Delete(name string) {
	s.init()

	s.Lock.Lock()
	defer s.Lock.Unlock()

	delete(s.Sessions, name)
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

// Sync is a ResponseSync implementation that is used to store the response
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
