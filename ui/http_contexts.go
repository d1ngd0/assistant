package ui

import (
	"bytes"
	"io"
	"net/http"

	"github.com/d1ngd0/assistant/state"
	"golang.org/x/net/websocket"
)

func ServeContext(sessions *state.Sessions) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, req *http.Request) {
		name := req.PathValue("session")
		if name == "" {
			R404(rw, req)
			return
		}

		session := sessions.Fetch(name)
		websocket.Handler(HandleWSContexts(session)).ServeHTTP(rw, req)
	})
}

func HandleWSContexts(session *state.Session) websocket.Handler {
	return func(ws *websocket.Conn) {
		var index int
		for {
			for msg, x := range session.Contexts(index) {
				SendWS(ws, "msg.html", msg)
				index = x + 1
			}

			select {
			case <-ws.Request().Context().Done():
				return
			case <-session.Wait():
			}
		}
	}

}

func SendWS(ws *websocket.Conn, tmplName string, val any) {
	var buf bytes.Buffer
	err := tmpl.ExecuteTemplate(&buf, tmplName, val)
	if err != nil {
		buf.Reset()
		Log.Printf("Issue rendering output: %s", err)
		_ = tmpl.ExecuteTemplate(&buf, "error_msg.html", err)
	}

	// TODO: You need to add error logic here and whatnot
	// very fragile
	_, _ = io.Copy(ws, &buf)
}
