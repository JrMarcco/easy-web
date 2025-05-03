package easyweb

import "context"

type TemplateEngine interface {
	Render(ctx context.Context, tplName string, data any) ([]byte, error)
}
