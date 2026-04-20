package cli

import (
	"errors"
	"fmt"
)

// ErrUser is the sentinel that marks errors caused by bad user input.
// main.go maps errors.Is(err, ErrUser) to exit code 1; anything else
// becomes exit code 2.
var ErrUser = errors.New("user error")

// UserErrorf returns an error that wraps ErrUser with a formatted
// message. Use it at call sites instead of hand-rolling fmt.Errorf
// so downstream errors.Is(err, ErrUser) keeps working.
func UserErrorf(format string, args ...any) error {
	return fmt.Errorf("%w: "+format, append([]any{ErrUser}, args...)...)
}
