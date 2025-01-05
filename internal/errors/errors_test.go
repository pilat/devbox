package errors

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrors(t *testing.T) {
	tests := []struct {
		errSample   func() error
		expectation string
	}{
		{
			func() error {
				e1a := errors.New("error 1a")
				e1b := fmt.Errorf("error 1b")

				e2 := fmt.Errorf("error 2: %w and %w", e1a, e1b)
				e3 := fmt.Errorf("error 3: %w", e2)

				return e3
			},
			`
error 3:
  error 2:
    error 1a
    error 1b
`,
		},
		{
			func() error {
				e1 := errors.New("error 1")
				e2 := fmt.Errorf("error 2: %w", e1)
				e3 := fmt.Errorf("error 3: %w", e2)

				return e3
			},
			`
error 3:
  error 2:
    error 1
`,
		},
		{
			func() error {
				e1 := errors.New("error 1")
				e2 := fmt.Errorf("error 2 + %w", e1)

				return e2
			},
			`
error 2 +
  error 1
`,
		},
	}

	for i, test := range tests {
		t.Run(fmt.Sprintf("Case %d", i), func(t *testing.T) {
			sampleErr := test.errSample()
			newErr := AsStacktrace(sampleErr)
			newErrAsString := newErr.Error()

			expectation := test.expectation
			expectation = expectation[1:] // remove first newline
			assert.Equal(t, expectation, newErrAsString)
		})
	}
}
