package easyweb

import (
	"bytes"
	"html/template"
)

type TemplateEngine interface {
	Render(tplName string, data any) ([]byte, error)
}
type GoTemplateEngine struct {
	T *template.Template
}

func (g *GoTemplateEngine) Render(tplName string, data any) ([]byte, error) {
	bs := &bytes.Buffer{}
	err := g.T.ExecuteTemplate(bs, tplName, data)
	return bs.Bytes(), err
}
