package ui

import (
	"bytes"
	"io"
	"net/http"

	"github.com/d1ngd0/assistant/state"
	"golang.org/x/net/websocket"
)

type Sessions struct {
	*state.Sessions
}

// NewUISession creates a new session object for use in the ui
func NewUISession(s *state.Session) *Session {
	return &Session{s}
}

type Session struct {
	*state.Session
}

func (s *Sessions) ServeContext(rw http.ResponseWriter, req *http.Request) {
	name := req.PathValue("session")
	if name == "" {
		R404(rw, req)
		return
	}

	session := Session{s.Fetch(name)}
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

func (s *Session) SendMsg(ws *websocket.Conn, msg state.Message) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, "msg.html", msg)
	if err != nil {
		buf.Reset()
		Log.Printf("Issue rendering output: %s", err)
		_ = tmpl.ExecuteTemplate(&buf, "error_msg.html", err)
	}

	// TODO: You need to add error logic here and whatnot
	// very fragile
	_, _ = io.Copy(ws, &buf)
}
