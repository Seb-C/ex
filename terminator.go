package ex

import (
	"errors"
	"fmt"
	"os"
	"runtime"
)

var onGarbageCollectUnclosed = func(err error) {
	fmt.Fprintln(os.Stderr, err)
}

func OnGarbageCollectUnclosed(handler func(error)) {
	if handler == nil {
		panic("Cannot set a nil garbage collect unclosed handler")
	}

	onGarbageCollectUnclosed = handler
}

type deferred struct {
	callback func() error
	fromFile string
	fromLine int
	isClosed bool
}

func (deferred *deferred) String() string {
	if deferred.fromFile == "" {
		return "deferred close"
	}

	return fmt.Sprintf(
		"deferred close initiated by %s:%d",
		deferred.fromFile,
		deferred.fromLine,
	)
}

func (deferred *deferred) finalizer() {
	if deferred.isClosed {
		return
	}

	onGarbageCollectUnclosed(fmt.Errorf(
		"The Close method of the terminator containing a %s was never called before being garbage collected.",
		deferred.String(),
	))
}

type Terminator struct {
	defers []*deferred // Needs to be a pointer for SetFinalizer
}

// The terminator is aimed to be included without a pointer,
// however we need one only when deferring stuff
func (terminator *Terminator) Defer(callback func() error) {
	if callback == nil {
		panic("A deferred close function cannot be nil")
	}

	// We do not handle the ok parameter here because we don't want to
	// fail this operation, and there is nothing much we can do if the
	// runtime information cannot be retrieved, so it can stay empty.
	_, file, line, _ := runtime.Caller(1)

	deferredClose := &deferred{
		callback: callback,
		fromFile: file,
		fromLine: line,
		isClosed: false,
	}
	terminator.defers = append(terminator.defers, deferredClose)

	runtime.SetFinalizer(deferredClose, (*deferred).finalizer)
}

func (terminator *Terminator) Close() error {
	var errs []error

	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(terminator.defers) - 1; callbackIndex >= 0; callbackIndex-- {
		deferred := terminator.defers[callbackIndex]
		if err := deferred.callback(); err != nil {
			errs = append(errs, fmt.Errorf("%s, caused by %w", deferred.String(), err))
		}
		deferred.isClosed = true
	}

	return errors.Join(errs...)
}
