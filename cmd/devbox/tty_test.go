package main

import (
	"os"
	"testing"
)

func TestIsTTYAvailable(t *testing.T) {
	devNull, err := os.Open(os.DevNull)
	if err != nil {
		t.Fatalf("failed to open %s: %v", os.DevNull, err)
	}
	t.Cleanup(func() { devNull.Close() })

	tmpFile, err := os.CreateTemp(t.TempDir(), "tty")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	t.Cleanup(func() { tmpFile.Close() })

	pipeR, pipeW, err := os.Pipe()
	if err != nil {
		t.Fatalf("failed to create pipe: %v", err)
	}
	t.Cleanup(func() {
		pipeR.Close()
		pipeW.Close()
	})

	tests := []struct {
		name string
		file *os.File
		want bool
	}{
		{name: "dev null is not a terminal", file: devNull, want: false},
		{name: "regular file is not a terminal", file: tmpFile, want: false},
		{name: "pipe is not a terminal", file: pipeR, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isTTYAvailable(tt.file); got != tt.want {
				t.Errorf("isTTYAvailable() = %v, want %v", got, tt.want)
			}
		})
	}
}
