package ex_test

import (
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
}
