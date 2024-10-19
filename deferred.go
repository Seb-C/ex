package ex

import (
	"fmt"
)

type deferred struct {
	callback           func() error
	fromFile           string
	fromLine           int
	gcUnclosedDetector *gcUnclosedDetector
}

func newDeferred(
	callback func() error,
	fromFile string,
	fromLine int,
) *deferred {
	deferred := &deferred{
		callback: callback,
		fromFile: fromFile,
		fromLine: fromLine,
	}
	deferred.gcUnclosedDetector = newGCUnclosedDetector(deferred.String())

	return deferred
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
