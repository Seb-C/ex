package _util

type Destructor struct {
	callbacks []func()
}

// The destructor is aimed to be included without a pointer,
// however we need one only when deferring stuff
func (destructor *Destructor) Defer(callback func()) {
	destructor.callbacks = append(destructor.callbacks, callback)
}

func (destructor *Destructor) Close() {
	// The callbacks must be destroyed from the most recent to the oldest
	for callbackIndex := len(destructor.callbacks) - 1; callbackIndex >= 0; callbackIndex-- {
		destructor.callbacks[callbackIndex]()
	}
}
