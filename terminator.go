package ex

type Terminator struct {
	callbacks []func()
}

// The terminator is aimed to be included without a pointer,
// however we need one only when deferring stuff
func (terminator *Terminator) Defer(callback func()) {
	terminator.callbacks = append(terminator.callbacks, callback)
}

func (terminator *Terminator) Close() {
	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(terminator.callbacks) - 1; callbackIndex >= 0; callbackIndex-- {
		terminator.callbacks[callbackIndex]()
	}
}
