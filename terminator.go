package ex

import "errors"

type Terminator struct {
	callbacks []func() error
}

// The terminator is aimed to be included without a pointer,
// however we need one only when deferring stuff
func (terminator *Terminator) Defer(callback func() error) {
	terminator.callbacks = append(terminator.callbacks, callback)
}

func (terminator *Terminator) Close() error {
	var errs []error

	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(terminator.callbacks) - 1; callbackIndex >= 0; callbackIndex-- {
		if err := terminator.callbacks[callbackIndex](); err != nil {
			errs = append(errs, err)
		}
	}

	return errors.Join(errs...)
}
