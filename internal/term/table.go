package term

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"golang.org/x/term"
)

type Table interface {
	Write()
}

type tableSvc struct {
	t table.Writer
}

var _ Table = &tableSvc{}

func NewTable(fields ...string) *tableSvc {
	w, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		w = 80
	}
	if w > 160 {
		w = 160
	}

	t := table.NewWriter()
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = true

	rows := make(table.Row, len(fields))
	for i, f := range fields {
		rows[i] = f
	}

	t.AppendHeader(rows)

	wm := w / len(fields)

	configs := make([]table.ColumnConfig, len(fields))
	for i := range configs {
		configs[i] = table.ColumnConfig{
			Number:      i + 1,
			AlignHeader: text.AlignCenter,
			AutoMerge:   true,
			WidthMax:    wm - 3,
			WidthMin:    wm - 3,
		}
	}

	t.SetColumnConfigs(configs)

	return &tableSvc{
		t: t,
	}
}

func (t *tableSvc) AppendRow(fields ...any) {
	t.t.AppendRow(fields)
}

func (t *tableSvc) Write() {
	t.t.SortBy([]table.SortBy{
		{Name: "Name", Mode: table.Asc},
	})
	fmt.Println(t.t.Render())
}
