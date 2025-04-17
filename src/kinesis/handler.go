package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/dynamo"
	"iss-telemetry-analyzer/src/sagemaker"
	"iss-telemetry-analyzer/src/types"
	"iss-telemetry-analyzer/src/utils"
	"iss-telemetry-analyzer/src/websocket"
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
	fmt.Printf("Received Kinesis record: %s\n", dataBytes)

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

		result := dynamo.StoreAnomalyScore(anomalyScore)
		anomalyLevel := utils.ComputeAnomalyLevel(result.Score, result.StandardDeviation, result.Average)

		if result.Error != nil {
			fmt.Printf("Error storing anomaly score: %v\n", result.Error)
		} else {
			fmt.Printf("Score: %f, Average: %f, Standard Deviation: %f, Anomaly Level: %s\n", result.Score, result.Average, result.StandardDeviation, anomalyLevel)
		}

		err := dynamo.StoreTelemetryData(anomalyScore, currentPressureValue, currentTemperatureValue, currentFlowRateValue, anomalyLevel)

		if err != nil {
			fmt.Println("ERRR ", err)
		}

		response := struct {
			Timestamp       string  `json:"timestamp"`
			FlowRate        float64 `json:"flowrate"`
			Pressure        float64 `json:"pressure"`
			Temperature     float64 `json:"temperature"`
			FlowChangeRate  float64 `json:"flow_change_rate"`
			PressChangeRate float64 `json:"press_change_rate"`
			TempChangeRate  float64 `json:"temp_change_rate"`
			AnomalyScore    float64 `json:"anomaly_score"`
			AnomalyLevel    string  `json:"anomaly_level"`
		}{
			Timestamp:       currentFlowRateTimestamp, // Use the timestamp from the flowrate data
			FlowRate:        currentFlowRateValue,
			Pressure:        currentPressureValue,
			Temperature:     currentTemperatureValue,
			FlowChangeRate:  flowChangeRate,
			PressChangeRate: pressureChangeRate,
			TempChangeRate:  temperatureChangeRate,
			AnomalyScore:    anomalyScore,
			AnomalyLevel:    string(anomalyLevel),
		}

		responseBytes, err := json.Marshal(response)

		if err != nil {
			fmt.Printf("Error marshaling response: %v\n", err)
		}

		// Update values
		previousFlowrateTimestamp = currentFlowRateTimestamp
		previousFlowrateValue = currentFlowRateValue

		previousPressureTimestamp = currentPressureTimestamp
		previousPressureValue = currentPressureValue

		previousTemperatureTimestamp = currentTemperatureTimestamp
		previousTemperatureValue = currentTemperatureValue

		websocket.PostMessage(responseBytes)
	}

	return nil
}
