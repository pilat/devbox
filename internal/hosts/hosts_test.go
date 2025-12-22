package hosts

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSave(t *testing.T) {
	tt := []struct {
		name            string
		initialContent  string
		projectName     string
		entries         []string
		expectedContent string
	}{
		{
			"File with no devbox entries",
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost",
			"test-project",
			[]string{"127.0.0.2 testhost"},
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost\n# BEGIN: Devbox 'test-project' project\n127.0.0.2 testhost\n# END: Devbox: 'test-project' project\n",
		},
		{
			"File with devbox entries",
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost\n# BEGIN: Devbox 'test-project' project\n127.0.0.2 testhost\n# END: Devbox: 'test-project' project\n",
			"test-project",
			[]string{"127.0.0.3 testhost2"},
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost\n# BEGIN: Devbox 'test-project' project\n127.0.0.3 testhost2\n# END: Devbox: 'test-project' project\n",
		},
		{
			"File with different devbox entries",
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost\n# BEGIN: Devbox 'test-project-2' project\n127.0.0.2 testhost\n# END: Devbox: 'test-project-2' project\n",
			"test-project",
			[]string{"127.0.0.4 testhost3"},
			"127.0.0.1 localhost\n\n\n  # some comment\n111.111.111.111 somehost\n# BEGIN: Devbox 'test-project-2' project\n127.0.0.2 testhost\n# END: Devbox: 'test-project-2' project\n# BEGIN: Devbox 'test-project' project\n127.0.0.4 testhost3\n# END: Devbox: 'test-project' project\n",
		},
	}

	for _, tc := range tt {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			tempFile, err := os.CreateTemp("", "hosts-test")
			assert.NoError(t, err)
			defer func() { _ = os.Remove(tempFile.Name()) }()

			err = os.WriteFile(tempFile.Name(), []byte(tc.initialContent), 0644)
			assert.NoError(t, err)

			_, err = save(tempFile.Name(), tc.projectName, tc.entries)
			assert.NoError(t, err)

			content, err := os.ReadFile(tempFile.Name())
			assert.NoError(t, err)

			assert.Equal(t, tc.expectedContent, string(content))
		})
	}
}
