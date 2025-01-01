package utils

import (
	"os"
	"strings"
)

func IsColorSupported() bool {
	term := os.Getenv("TERM")
	if term == "" {
		return false
	}

	supportedTerms := []string{"xterm", "xterm-256color", "screen", "screen-256color", "tmux", "tmux-256color"}
	for _, t := range supportedTerms {
		if strings.Contains(term, t) {
			return true
		}
	}

	return false
}
