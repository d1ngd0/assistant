package ui

import (
	"bytes"
	"io"
	"iter"
	"net/http"
	"sync"

	"golang.org/x/net/websocket"
)

var ActiveSessions Sessions

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

func (s *Sessions) Fetch(name string) *Session {
	s.init()

	s.Lock.RLock()
	session, ok := s.Sessions[name]
	s.Lock.RUnlock()

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
	context []string
}

func newSession() *Session {
	s := &Session{
		event: make(chan struct{}),
	}
	return s
}

// Push places the latest message into the context
func (s *Session) Push(msg string) []string {
	s.lock.Lock()
	defer s.lock.Unlock()

	close(s.event)
	s.event = make(chan struct{})

	s.context = append(s.context, msg)
	return s.context
}

// Sync is a ResponseSync implementation that is used to store the response
// in the context.
func (s *Session) Sync(msg string) error {
	_ = s.Push(msg)
	return nil
}

// Contexts returns an iterator which locks the session context until it is exhausted
func (s *Session) Contexts(from int) iter.Seq2[string, int] {
	s.lock.RLock()
	return func(yield func(msg string, index int) bool) {
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

func (s *Sessions) ServeContext(rw http.ResponseWriter, req *http.Request) {
	s.init()

	name := req.PathValue("session")
	if name == "" {
		R404(rw, req)
		return
	}

	session := s.Fetch(name)
	websocket.Handler(session.HandleConn).ServeHTTP(rw, req)
}

func (s *Session) HandleConn(ws *websocket.Conn) {
	var index int
	for {
		for msg, x := range s.Contexts(index) {
			s.SendMsg(ws, msg)
			index = x + 1
		}

		select {
		case <-ws.Request().Context().Done():
			return
		case <-s.Wait():
		}
	}

}

func (s *Session) SendMsg(ws *websocket.Conn, msg string) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "msg.html", msg)
	if err != nil {
		return
	}

	// TODO: You need to add error logic here and whatnot
	// very fragile
	_, _ = io.Copy(ws, &buf)
}
