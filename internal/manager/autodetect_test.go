package manager

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNormalizeRemoteURL(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			"simple https url",
			"https://github.com/org/repo",
			"github.com/org/repo",
		},
		{
			"https url with .git",
			"https://github.com/org/repo.git",
			"github.com/org/repo",
		},
		{
			"https url with username and password",
			"https://user:pass@github.com/org/repo.git",
			"github.com/org/repo",
		},
		{
			"https url with token",
			"https://token@github.com/org/repo",
			"github.com/org/repo",
		},
		{
			"ssh url with git@",
			"git@github.com:org/repo.git",
			"github.com/org/repo",
		},
		{
			"ssh url with protocol",
			"ssh://git@github.com/org/repo.git",
			"github.com/org/repo",
		},
		{
			"mixed case url",
			"HTTPS://GitHub.com/Org/Repo.git",
			"github.com/org/repo",
		},
		{
			"gitlab ssh url",
			"git@gitlab.com:group/subgroup/repo.git",
			"gitlab.com/group/subgroup/repo",
		},
		{
			"bitbucket url with username",
			"https://user@bitbucket.org/org/repo.git",
			"bitbucket.org/org/repo",
		},
		{
			"url with port",
			"ssh://git@github.com:22/org/repo.git",
			"github.com:22/org/repo",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRemoteURL(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
