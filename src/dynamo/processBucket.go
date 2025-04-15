package dynamo

import (
	"fmt"
	"iss-telemetry-analyzer/src/types"
	"strconv"
)

var lastValues = map[string]float64{"FLOWRATE": 0, "PRESSURE": 0, "TEMPERATURE": 0}
var chunkSizeSeconds = 5

func ProcessBucket(bucketKey string, data []types.TelemetryData) *types.ProcessedData {
	// Initialize with last known values
	current := map[string]float64{
		"FLOWRATE":    lastValues["FLOWRATE"],
		"PRESSURE":    lastValues["PRESSURE"],
		"TEMPERATURE": lastValues["TEMPERATURE"],
	}

	// Update with latest values in bucket
	for _, d := range data {
		value, err := strconv.ParseFloat(d.Value, 64)
		if err != nil {
			fmt.Printf("Error parsing %s value %s: %v\n", d.Name, d.Value, err)
			continue
		}
		current[d.Name] = value
	}

	// Calculate rates of change (over 5 seconds)
	processed := types.ProcessedData{
		Timestamp:       bucketKey,
		FLOWRATE:        current["FLOWRATE"],
		PRESSURE:        current["PRESSURE"],
		TEMPERATURE:     current["TEMPERATURE"],
		FlowChangeRate:  (current["FLOWRATE"] - lastValues["FLOWRATE"]) / float64(chunkSizeSeconds),
		PressChangeRate: (current["PRESSURE"] - lastValues["PRESSURE"]) / float64(chunkSizeSeconds),
		TempChangeRate:  (current["TEMPERATURE"] - lastValues["TEMPERATURE"]) / float64(chunkSizeSeconds),
	}

	// Update last values for next bucket
	lastValues["FLOWRATE"] = current["FLOWRATE"]
	lastValues["PRESSURE"] = current["PRESSURE"]
	lastValues["TEMPERATURE"] = current["TEMPERATURE"]

	return &processed
}
