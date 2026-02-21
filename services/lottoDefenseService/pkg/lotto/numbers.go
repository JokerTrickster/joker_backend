package lotto

import (
	"math/rand"
	"sort"
)

// MinNumber is the minimum lotto number (inclusive)
const MinNumber = 1

// MaxNumber is the maximum lotto number (inclusive)
const MaxNumber = 45

// DrawCount is the number of balls drawn
const DrawCount = 6

// Draw generates 6 unique random numbers between 1 and 45 (inclusive), sorted ascending
func Draw() [6]int {
	used := make(map[int]bool)
	var out [6]int
	for i := 0; i < DrawCount; i++ {
		for {
			n := MinNumber + rand.Intn(MaxNumber-MinNumber+1)
			if !used[n] {
				used[n] = true
				out[i] = n
				break
			}
		}
	}
	sort.Ints(out[:])
	return out
}
