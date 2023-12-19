package ex

import (
	"errors"
	"fmt"
	"runtime"
)

// Terminator should be embedded into any struct that is responsible
// for resources that has to be closed.
type Terminator struct {
	defers []*deferred // Needs to be a pointer for SetFinalizer
}

// Defer mimicks the defer keyword, but throughout the life cycle of your struct.
//
// Everything deferred by it will be automatically called by the Close method.
func (terminator *Terminator) Defer(callback func() error) {
	if callback == nil {
		panic("A deferred close function cannot be nil")
	}

	// We do not handle the ok parameter here because we don't want to
	// fail this operation, and there is nothing much we can do if the
	// runtime information cannot be retrieved, so it can stay empty.
	_, file, line, _ := runtime.Caller(1)

	deferredClose := newDeferred(callback, file, line)
	terminator.defers = append(terminator.defers, deferredClose)
}

// Close will execute all the deferred functions that were previously passed to Defer.
//
// It uses the same order than the defer keyword: from the last Defer to the first.
//
// All of the deferred functions are always executed, even if one of them fails.
//
// If many of them fails, all the related errors will be returned using errors.Join,
// along with a trace.
func (terminator *Terminator) Close() error {
	var errs []error

	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(terminator.defers) - 1; callbackIndex >= 0; callbackIndex-- {
		deferred := terminator.defers[callbackIndex]
		if err := deferred.callback(); err != nil {
			errs = append(errs, fmt.Errorf("%s, caused by %w", deferred.String(), err))
		}
		deferred.gcUnclosedDetector.isClosed = true
	}
	terminator.defers = nil

	return errors.Join(errs...)
}
