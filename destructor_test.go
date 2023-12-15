package ex

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDestructor(t *testing.T) {
	t.Run("basic non-pointer", func(t *testing.T) {
		t.Run("non-pointer", func(t *testing.T) {
			obj := struct{ Destructor }{}

			counterA := 0
			obj.Defer(func() {
				counterA++
			})

			counterB := 0
			obj.Defer(func() {
				counterB++
			})

			obj.Close()

			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
		t.Run("pointer", func(t *testing.T) {
			obj := &struct{ Destructor }{}

			counterA := 0
			obj.Defer(func() {
				counterA++
			})

			counterB := 0
			obj.Defer(func() {
				counterB++
			})

			obj.Close()

			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
	})

	t.Run("embedded", func(t *testing.T) {
		t.Run("non-pointer", func(t *testing.T) {
			type typeA struct {
				Destructor
			}
			type typeB struct {
				Destructor
				a typeA
			}

			objA := typeA{}
			objB := typeB{a: objA}

			objB.Defer(objA.Close)

			counterA := 0
			objA.Defer(func() {
				counterA++
			})

			counterB := 0
			objB.Defer(func() {
				counterB++
			})

			objB.Close()

			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
		t.Run("pointer", func(t *testing.T) {
			type typeA struct {
				Destructor
			}
			type typeB struct {
				Destructor
				a *typeA
			}

			objA := &typeA{}
			objB := &typeB{a: objA}

			objB.Defer(objA.Close)

			counterA := 0
			objA.Defer(func() {
				counterA++
			})

			counterB := 0
			objB.Defer(func() {
				counterB++
			})

			objB.Close()

			assert.Equal(t, 1, counterA)
			assert.Equal(t, 1, counterB)
		})
	})
}
