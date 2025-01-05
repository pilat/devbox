package table

import (
	"fmt"
	"os"

	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"golang.org/x/term"
)

type tableSvc struct {
	table.Writer
	fields    []string
	compact   bool
	sortBySet bool
}

func New(fields ...string) *tableSvc {
	return &tableSvc{
		Writer:    table.NewWriter(),
		fields:    fields,
		compact:   false,
		sortBySet: false,
	}
}

func (t *tableSvc) AppendRow(fields ...any) {
	t.Writer.AppendRow(fields)
}

func (t *tableSvc) Compact() {
	t.compact = true
}

func (t *tableSvc) SortBy(sortBy []SortBy) {
	t.sortBySet = true
	t.Writer.SortBy(sortBy)
}

func (t *tableSvc) Render() {
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = !t.compact

	rows := make(table.Row, len(t.fields))
	for i, f := range t.fields {
		rows[i] = f
	}

	t.AppendHeader(rows)

	w := t.getTerminalWidth()
	wm := w / len(t.fields)

	configs := make([]table.ColumnConfig, len(t.fields))
	for i := range configs {
		c := table.ColumnConfig{
			Number:      i + 1,
			AlignHeader: text.AlignCenter,
		}

		if !t.compact {
			c.AutoMerge = true
			c.WidthMax = wm - 3
			c.WidthMin = wm - 3
		}

		configs[i] = c
	}

	t.SetColumnConfigs(configs)

	if !t.sortBySet {
		t.Writer.SortBy([]table.SortBy{
			{Name: "Name", Mode: table.Asc},
		})
	}

	fmt.Println(t.Writer.Render())
}

func (t *tableSvc) getTerminalWidth() int {
	w, _, err := term.GetSize(int(os.Stdin.Fd()))
	if err != nil {
		w = 80
	}
	if w > 160 {
		w = 160
	}

	return w
}
