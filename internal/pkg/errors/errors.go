package errors

import (
	"errors"
	"strings"
)

func AsStacktrace(err error) error {
	return errors.New(asStacktrace(err, 0))
}

func asStacktrace(err error, padding int) string {
	out := ""
	var errs []error
	if err, ok := err.(interface{ Unwrap() error }); ok {
		errs = []error{err.Unwrap()}
	}

	if err, ok := err.(interface{ Unwrap() []error }); ok {
		errs = err.Unwrap()
	}

	errMsg := err.Error()
	if len(errs) > 0 {
		firstInnerErrMsg := errs[0].Error()
		if pos := strings.Index(errMsg, firstInnerErrMsg); pos != -1 {
			errMsg = errMsg[:pos]
		}

		errMsg = strings.TrimSpace(errMsg)
	}

	out += strings.Repeat(" ", padding) + errMsg + "\n"
	for _, err := range errs {
		out += asStacktrace(err, padding+2)
	}

	return out
}
