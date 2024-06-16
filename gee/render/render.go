package render

import "net/http"

type Render interface {
	Render(w http.ResponseWriter, code int) error
	WriteContentType(w http.ResponseWriter)
}

func writeContentType(w http.ResponseWriter, contentType string) {
	w.Header().Set("Content-Type", contentType)
}
