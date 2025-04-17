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
	}

	return valueDiff / timeDiff
}
