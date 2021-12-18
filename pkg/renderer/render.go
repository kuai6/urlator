package renderer

import (
	"sort"

	"github.com/dustin/go-humanize"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/kuai6/urlator/pkg/runner"
)

type renderer struct {
}

func NewRenderer() *renderer {
	return &renderer{}
}

func (r *renderer) Render(res []runner.Result) string {

	sort.Slice(res, func(i, j int) bool {
		return res[i].Size > res[j].Size
	})

	t := table.NewWriter()
	t.AppendHeader(table.Row{"#", "URL", "Result"})
	for i, row := range res {

		value := humanize.Bytes(row.Size)
		if row.Err != nil {
			value = row.Err.Error()
		}
		t.AppendRow(table.Row{i, row.Url, value})
	}

	return t.Render()
}
