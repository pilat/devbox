package app

import (
	"fmt"
	"os"
	"time"

	"github.com/jedib0t/go-pretty/v6/progress"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"

	"golang.org/x/term"
)

func createProgress() progress.Writer {
	pw := progress.NewWriter()
	pw.SetMessageLength(100)
	pw.SetStyle(progress.StyleDefault)
	pw.SetTrackerPosition(progress.PositionRight)
	pw.SetUpdateFrequency(time.Millisecond * 100)
	pw.Style().Colors = progress.StyleColorsExample
	pw.Style().Visibility.ETA = false
	pw.Style().Visibility.ETAOverall = false
	pw.Style().Visibility.Percentage = false
	pw.Style().Visibility.Speed = false
	pw.Style().Visibility.SpeedOverall = false
	pw.Style().Visibility.Value = false

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

func stopProgress(pw progress.Writer) {
	time.Sleep(200 * time.Millisecond)
	pw.Stop()
	time.Sleep(200 * time.Millisecond)
}

func makeTable(fields ...string) table.Writer {
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

	return t
}

func renderTable(t table.Writer) {
	t.SortBy([]table.SortBy{
		{Name: "Name", Mode: table.Asc},
	})
	fmt.Println(t.Render())
}
