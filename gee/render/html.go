package render

import (
	"html/template"
	"net/http"

	"github.com/gee-coder/gee/internal/bytesconv"
)

type HTML struct {
	Name       string
	Template   *template.Template
	IsTemplate bool
	Data       any
}
type HTMLRender struct {
	Template *template.Template
}

func (h *HTML) Render(w http.ResponseWriter, code int) error {
	h.WriteContentType(w)
	w.WriteHeader(code)
	if h.IsTemplate {
		err := h.Template.ExecuteTemplate(w, h.Name, h.Data)
		return err
	}
	_, err := w.Write(bytesconv.StringToBytes(h.Data.(string)))
	return err
}

func (h *HTML) WriteContentType(w http.ResponseWriter) {
	writeContentType(w, "text/html; charset=utf-8")
}
