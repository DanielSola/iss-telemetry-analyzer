package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/dynamo"
	"iss-telemetry-analyzer/src/sagemaker"
	"iss-telemetry-analyzer/src/types"
	"iss-telemetry-analyzer/src/utils"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
)

var currentTemperatureValue float64
var currentTemperatureTimestamp string
var previousTemperatureValue float64
var previousTemperatureTimestamp string

// Pressure
var currentPressureValue float64
var currentPressureTimestamp string
var previousPressureValue float64
var previousPressureTimestamp string

// Flowrate
var currentFlowRateValue float64
var currentFlowRateTimestamp string
var previousFlowrateValue float64
var previousFlowrateTimestamp string

func Handler(ctx context.Context, kinesisEvent events.KinesisEvent) error {

	// Process each Kinesis record
	record := kinesisEvent.Records[0]
	dataBytes := record.Kinesis.Data

	var telemetryData types.TelemetryData

	if err := json.Unmarshal(dataBytes, &telemetryData); err != nil {
		fmt.Printf("Cannot read Kinesis telemetry data: %v\n", err)
	}

	if telemetryData.Name == "FLOWRATE" {
		flowrate, err := strconv.ParseFloat(telemetryData.Value, 64)

		if err != nil {
			fmt.Printf("Error parsing %s value %s: %v\n", telemetryData.Name, telemetryData.Value, err)
		}

		currentFlowRateValue = flowrate
		currentFlowRateTimestamp = telemetryData.Timestamp
	}

	if telemetryData.Name == "TEMPERATURE" {
		temperature, err := strconv.ParseFloat(telemetryData.Value, 64)

		if err != nil {
			fmt.Printf("Error parsing %s value %s: %v\n", telemetryData.Name, telemetryData.Value, err)
		}

		currentTemperatureValue = temperature
		currentTemperatureTimestamp = telemetryData.Timestamp
	}

	if telemetryData.Name == "PRESSURE" {
		pressure, err := strconv.ParseFloat(telemetryData.Value, 64)

		if err != nil {
			fmt.Printf("Error parsing %s value %s: %v\n", telemetryData.Name, telemetryData.Value, err)
		}

		currentPressureValue = pressure
		currentPressureTimestamp = telemetryData.Timestamp
	}

	if currentFlowRateValue != 0.0 && currentTemperatureValue != 0.0 && currentPressureValue != 0.0 {

		flowChangeRate := utils.GetChangeRate(currentFlowRateValue, previousFlowrateValue, currentFlowRateTimestamp, previousFlowrateTimestamp)
		temperatureChangeRate := utils.GetChangeRate(currentTemperatureValue, previousTemperatureValue, currentTemperatureTimestamp, previousTemperatureTimestamp)
		pressureChangeRate := utils.GetChangeRate(currentPressureValue, previousPressureValue, currentPressureTimestamp, previousPressureTimestamp)

		features := []float64{
			currentFlowRateValue, currentPressureValue, currentPressureValue,
			flowChangeRate, pressureChangeRate, temperatureChangeRate,
		}

		var anomalyScore = sagemaker.Predict(features)

		scoreResult := dynamo.StoreAnomalyScore(anomalyScore)

		if scoreResult.Error != nil {
			fmt.Println("STORE ERRORS: ", scoreResult.Error)
		}

		// Log the telemetry data for querying in Grafana (structured JSON format)
		logData := map[string]interface{}{
			"timestamp":         currentFlowRateTimestamp,
			"flowrate":          currentFlowRateValue,
			"pressure":          currentPressureValue,
			"temperature":       currentTemperatureValue,
			"flow_change_rate":  flowChangeRate,
			"press_change_rate": pressureChangeRate,
			"temp_change_rate":  temperatureChangeRate,
			"anomaly_score":     anomalyScore,
			//"anomaly_level":     anomalyLevel,
			"log_type":            "telemetry_data", // A tag for identifying the type of log entry
			"last_hour_avg_score": scoreResult.Average,
		}

		logDataBytes, err := json.Marshal(logData)

		if err != nil {
			fmt.Printf("Error marshaling log data: %v\n", err)
		} else {
			// Print log data as JSON for CloudWatch and Grafana to query
			fmt.Println(string(logDataBytes))
		}

		// Update values
		previousFlowrateTimestamp = currentFlowRateTimestamp
		previousFlowrateValue = currentFlowRateValue

		previousPressureTimestamp = currentPressureTimestamp
		previousPressureValue = currentPressureValue

		previousTemperatureTimestamp = currentTemperatureTimestamp
		previousTemperatureValue = currentTemperatureValue

	}

	return nil
}
