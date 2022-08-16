package binutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_GetDiscreteStatuses(t *testing.T) {

	t.Run("type alias", func(t *testing.T) {

		type status int32

		var s1 status = 3

		resp := GetDiscreteStatuses(s1, 64)
		assert.Equal(t, []status{1, 2}, resp)
	})

	t.Run("empty", func(t *testing.T) {
		resp := GetDiscreteStatuses(0, 64)
		assert.Equal(t, resp, []int{})
	})

	t.Run("int over max", func(t *testing.T) {
		resp := GetDiscreteStatuses(111, 8)
		assert.Equal(t, resp, []int{1, 2, 4})
	})

	t.Run("large", func(t *testing.T) {
		resp := GetDiscreteStatuses(127, 512)
		assert.Equal(t, []int{1, 2, 4, 8, 16, 32, 64}, resp)
	})
}
