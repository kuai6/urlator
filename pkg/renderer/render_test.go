package renderer

import (
	"testing"

	"github.com/kuai6/urlator/pkg/runner"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func Test_renderer_Render(t *testing.T) {

	r := &renderer{}
	got := r.Render([]runner.Result{
		{
			Url:  "https://foo.bar",
			Size: 10,
			Err:  nil,
		},
		{
			Url:  "https://fuzz.buzz",
			Size: 100,
			Err:  nil,
		},
		{
			Url:  "https://ya.ya",
			Size: 0,
			Err:  errors.New("some error"),
		},
	})

	expectedOutput := `+---+-------------------+------------+
| # | URL               | RESULT     |
+---+-------------------+------------+
| 0 | https://fuzz.buzz | 100 B      |
| 1 | https://foo.bar   | 10 B       |
| 2 | https://ya.ya     | some error |
+---+-------------------+------------+`

	assert.Equal(t, expectedOutput, got)
}
