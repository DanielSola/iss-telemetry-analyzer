package utils

import (
	"fmt"
	"time"
)

// GetTimeDiff calculates the difference in seconds between two ISO 8601 timestamps.
func GetTimeDiff(timestamp1, timestamp2 string) (float64, error) {
	// Parse the first timestamp
	t1, err := time.Parse(time.RFC3339, timestamp1)
	if err != nil {
		return 0, fmt.Errorf("error parsing timestamp1: %v", err)
	}

	// Parse the second timestamp
	t2, err := time.Parse(time.RFC3339, timestamp2)
	if err != nil {
		return 0, fmt.Errorf("error parsing timestamp2: %v", err)
	}

	// Calculate the difference in seconds
	duration := t2.Sub(t1).Seconds()

	return duration, nil
}
