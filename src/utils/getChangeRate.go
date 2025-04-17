package utils

import "fmt"

func GetChangeRate(currentValue float64, previousValue float64, currentTimestamp string, previousTimestamp string) float64 {

	if previousTimestamp == "" {
		return 0
	}

	valueDiff := (currentValue - previousValue)
	timeDiff, err := GetTimeDiff(currentTimestamp, previousTimestamp)

	if err != nil {
		fmt.Printf("Error calculating time diff between: %s and %s: %v\n", currentTimestamp, previousTimestamp, err)
		return 0 // Return 0 if there's an error in time difference calculation
	}

	if timeDiff == 0 {
		return 0 // Avoid division by zero
	}

	return valueDiff / timeDiff
}
