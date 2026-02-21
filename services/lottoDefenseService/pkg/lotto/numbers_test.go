package lotto

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDraw_ReturnsSixNumbers(t *testing.T) {
	for i := 0; i < 20; i++ {
		nums := Draw()
		assert.Len(t, nums, 6, "Draw must return exactly 6 numbers")
	}
}

func TestDraw_NumbersInRange(t *testing.T) {
	for i := 0; i < 50; i++ {
		nums := Draw()
		for _, n := range nums {
			assert.GreaterOrEqual(t, n, MinNumber, "each number must be >= MinNumber")
			assert.LessOrEqual(t, n, MaxNumber, "each number must be <= MaxNumber")
		}
	}
}

func TestDraw_NumbersUnique(t *testing.T) {
	for i := 0; i < 50; i++ {
		nums := Draw()
		seen := make(map[int]bool)
		for _, n := range nums {
			require.False(t, seen[n], "numbers must be unique, duplicate found: %d", n)
			seen[n] = true
		}
	}
}

func TestDraw_NumbersSortedAscending(t *testing.T) {
	for i := 0; i < 20; i++ {
		nums := Draw()
		for j := 1; j < len(nums); j++ {
			assert.GreaterOrEqual(t, nums[j], nums[j-1], "numbers must be sorted ascending")
		}
	}
}
