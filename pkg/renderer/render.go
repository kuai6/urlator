package renderer

import "github.com/kuai6/urlator/pkg/runner"

type renderer struct {
}

func NewRenderer() *renderer {
	return &renderer{}
}

func (r *renderer) Render([]runner.Result) string {
	return ""
}
