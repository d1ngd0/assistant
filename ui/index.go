package ui

import (
	"bytes"
	"net/http"
)

func ServeIndex() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var b bytes.Buffer
		err := tmpl.ExecuteTemplate(&b, "index.html", nil)
		if err != nil {
			R500(w, req, err)
			return
		}

		w.Write(b.Bytes())
	})
}
