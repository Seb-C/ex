package ex

import (
	"fmt"
	"os"
	"runtime"
)

var onGarbageCollectUnclosed = func(err error) {
	fmt.Fprintln(os.Stderr, err)
}

// OnGarbageCollectUnclosed changes what happens when a Terminator
// gets garbage collected without the Close method having been called.
//
// The default behaviour is to write an error message to stderr.
// Use this for example if you prefer to panic or pass it to your log
// system instead of stderr.
func OnGarbageCollectUnclosed(handler func(error)) {
	if handler == nil {
		panic("Cannot set a nil garbage collect unclosed handler")
	}

	onGarbageCollectUnclosed = handler
}

type gcUnclosedDetector struct{
	description string
	isClosed bool
}

func newGCUnclosedDetector(description string) *gcUnclosedDetector {
	detector := &gcUnclosedDetector{
		description: description,
		isClosed: false,
	}

	runtime.SetFinalizer(detector, (*gcUnclosedDetector).finalizer)

	return detector
}

func (detector *gcUnclosedDetector) finalizer() {
	if detector.isClosed {
		return
	}

	onGarbageCollectUnclosed(fmt.Errorf(
		"The Close method of the terminator containing a %s was never called before being garbage collected.",
		detector.description,
	))
}
