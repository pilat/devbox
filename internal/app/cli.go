package app

import (
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
)

func createProgress() progress.Writer {
	pw := progress.NewWriter()
	pw.SetMessageLength(50)
	pw.SetSortBy(progress.SortByPercentDsc)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Speed = false
	pw.Style().Visibility.SpeedOverall = false
	pw.Style().Visibility.Tracker = false
	pw.Style().Visibility.Value = false

	pw.Style().Options.Separator = " | "
	pw.Style().Options.DoneString = "done"
	pw.Style().Options.ErrorString = "fail"

	go pw.Render()

	return pw
}

func addTracker(pw progress.Writer, message string, roc bool) *progress.Tracker {
	tracker := progress.Tracker{
		DeferStart:         true,
		Message:            message,
		RemoveOnCompletion: roc,
	}

	pw.AppendTracker(&tracker)

	return &tracker
}

func stopProgress(pw progress.Writer) func() {
	return func() {
		time.Sleep(110 * time.Millisecond)
		pw.Stop()
	}
}

func makeTable(fields ...string) table.Writer {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.SetStyle(table.StyleRounded)
	t.Style().Options.SeparateRows = true

	rows := make(table.Row, len(fields))
	for i, f := range fields {
		rows[i] = f
	}

	t.AppendHeader(rows)

	return t
}

func renderTable(t table.Writer, width ...int) {
	t.SortBy([]table.SortBy{
		{Name: "Name", Mode: table.Asc},
	})

	configs := make([]table.ColumnConfig, len(width))
	for i, w := range width {
		configs[i] = table.ColumnConfig{
			Number:      i + 1,
			AlignHeader: text.AlignCenter,
			AutoMerge:   true,
			WidthMin:    w,
			WidthMax:    w,
		}
	}
	t.SetColumnConfigs(configs)
	t.Render()
}
