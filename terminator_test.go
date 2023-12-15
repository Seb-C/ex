package ex_test

import (
	"errors"
	"regexp"
	"runtime"
	"testing"

	"github.com/Seb-C/ex"
	"github.com/stretchr/testify/assert"
)

func TestTerminator(t *testing.T) {
	t.Run("basic non-pointer", func(t *testing.T) {
		t.Run("non-pointer", func(t *testing.T) {
			obj := struct{ ex.Terminator }{}

			counterA := 0
			obj.Defer(func() error {
				counterA++
				return nil
			})

			counterB := 0
			obj.Defer(func() error {
				counterB++
				return nil
			})

			err := obj.Close()

			assert.Nil(t, err)
			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
		t.Run("pointer", func(t *testing.T) {
			obj := &struct{ ex.Terminator }{}

			counterA := 0
			obj.Defer(func() error {
				counterA++
				return nil
			})

			counterB := 0
			obj.Defer(func() error {
				counterB++
				return nil
			})

			err := obj.Close()

			assert.Nil(t, err)
			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
	})

	t.Run("embedded", func(t *testing.T) {
		t.Run("non-pointer", func(t *testing.T) {
			type typeA struct {
				ex.Terminator
			}
			type typeB struct {
				ex.Terminator
				a typeA
			}

			objA := typeA{}
			objB := typeB{a: objA}
			objB.Defer(objA.Close)

			counterA := 0
			objA.Defer(func() error {
				counterA++
				return nil
			})

			counterB := 0
			objB.Defer(func() error {
				counterB++
				return nil
			})

			err := objB.Close()

			assert.Nil(t, err)
			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
		t.Run("pointer", func(t *testing.T) {
			type typeA struct {
				ex.Terminator
			}
			type typeB struct {
				ex.Terminator
				a *typeA
			}

			objA := &typeA{}
			objB := &typeB{a: objA}

			objB.Defer(objA.Close)

			counterA := 0
			objA.Defer(func() error {
				counterA++
				return nil
			})

			counterB := 0
			objB.Defer(func() error {
				counterB++
				return nil
			})

			err := objB.Close()

			assert.Nil(t, err)
			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
	})

	t.Run("errors", func(t *testing.T) {
		t.Run("all returned with debug info", func(t *testing.T) {
			errA := errors.New("error A")
			errB := errors.New("error B")

			var terminator = ex.Terminator{}
			terminator.Defer(func() error {
				return errA
			})
			terminator.Defer(func() error {
				return errB
			})

			resultErr := terminator.Close()

			assert.NotNil(t, resultErr)
			assert.ErrorIs(t, resultErr, errA)
			assert.ErrorIs(t, resultErr, errB)
			assert.Contains(t, resultErr.Error(), "terminator_test.go")
		})
		t.Run("deep error text", func(t *testing.T) {
			type typeA struct {
				ex.Terminator
			}
			type typeB struct {
				ex.Terminator
				a typeA
			}
			type typeC struct {
				ex.Terminator
				b1 typeB
				b2 typeB
			}

			objA1 := typeA{}
			objA1.Defer(func() error {
				return errors.New("error X1")
			})
			objA1.Defer(func() error {
				return errors.New("error Y1")
			})

			objA2 := typeA{}
			objA2.Defer(func() error {
				return errors.New("error X2")
			})
			objA2.Defer(func() error {
				return errors.New("error Y2")
			})

			objB1 := typeB{a: objA1}
			objB1.Defer(objA1.Close)

			objB2 := typeB{a: objA2}
			objB2.Defer(objA2.Close)

			objC := typeC{b1: objB1, b2: objB2}
			objC.Defer(objB1.Close)
			objC.Defer(objB2.Close)

			err := objC.Close()

			assert.NotNil(t, err)
			assert.Regexp(t, regexp.MustCompile(`deferred close initiated by .*terminator_test.go:\d+, caused by deferred close initiated by .*terminator_test.go:\d+, caused by deferred close initiated by .*terminator_test.go:\d+, caused by error Y2`), err.Error())
			assert.Regexp(t, regexp.MustCompile(`deferred close initiated by /.*terminator_test.go:\d+, caused by deferred close initiated by .*/terminator_test.go:\d+, caused by deferred close initiated by .*/terminator_test.go:\d+, caused by error Y1`), err.Error())
		})
		t.Run("missing Close", func(t *testing.T) {
			type typeA struct {
				ex.Terminator
			}

			objA1 := typeA{}
			objA1.Defer(func() error {
				return nil
			})

			unclosedErrorCount := 0
			ex.OnGarbageCollectUnclosed(func(err error) {
				unclosedErrorCount++
				t.Log(err)
			})

			objA1 = typeA{} // Replacing the value so that it gets garbage collected
			runtime.GC()
			assert.Equal(t, 1, unclosedErrorCount)
		})
	})
}
