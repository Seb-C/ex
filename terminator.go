package ex

import (
	"errors"
	"fmt"
	"runtime"
)

type deferred struct {
	callback func() error
	fromFile string
	fromLine int
}

type Terminator struct {
	defers []deferred
}

// The terminator is aimed to be included without a pointer,
// however we need one only when deferring stuff
func (terminator *Terminator) Defer(callback func() error) {
	// We do not handle the ok parameter here because we don't want to
	// fail this operation, and there is nothing much we can do if the
	// runtime information cannot be retrieved, so it can stay empty.
	_, file, line, _ := runtime.Caller(1)

	terminator.defers = append(terminator.defers, deferred{
		callback: callback,
		fromFile: file,
		fromLine: line,
	})
}

func (terminator *Terminator) Close() error {
	var errs []error

	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(terminator.defers) - 1; callbackIndex >= 0; callbackIndex-- {
		deferred := terminator.defers[callbackIndex]
		if err := deferred.callback(); err != nil {
			var message = "deferred close"
			if deferred.fromFile != "" {
				message = fmt.Sprintf(
					"deferred close initiated by %s:%d",
					deferred.fromFile,
					deferred.fromLine,
				)
			}

			errs = append(errs, fmt.Errorf("%s, caused by %w", message, err))
		}
	}

	return errors.Join(errs...)
}
