package main

import (
	"fmt"
)

type total struct {
	id       int
	distance float64
}

// Formats a total as a string.
func (t total) String() string {
	return fmt.Sprintf("Traveler %d traveled %.2f miles", t.id, t.distance)
}
