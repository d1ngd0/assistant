package ui

import (
	"bytes"
	"net/http"
)

func ServeIndex() http.Handler {
	type context struct {
		Session string
	}

	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		session := req.PathValue("session")

		var b bytes.Buffer
		err := tmpl.ExecuteTemplate(&b, "index.html", context{Session: session})
		if err != nil {
			R500(w, req, err)
			return
		}

		_, _ = w.Write(b.Bytes())
	})
}
