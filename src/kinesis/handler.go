package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/sagemaker"
	"iss-telemetry-analyzer/src/websocket"
	"strconv"
	"sync"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

type TelemetryData struct {
	Name      string `json:"name"`      // e.g., "FLOWRATE", "PRESSURE", "TEMPERATURE"
	Value     string `json:"value"`     // String value to parse
	Timestamp string `json:"timestamp"` // ISO timestamp
}

type ProcessedData struct {
	Timestamp       string  `json:"timestamp"` // Start of 5s bucket
	FLOWRATE        float64 `json:"flowrate"`
	PRESSURE        float64 `json:"pressure"`
	TEMPERATURE     float64 `json:"temperature"`
	FlowChangeRate  float64 `json:"flow_change_rate"`
	PressChangeRate float64 `json:"press_change_rate"`
	TempChangeRate  float64 `json:"temp_change_rate"`
	AnomalyScore    float64 `json:"anomaly_score"`
}

// Buffer for 5-second windows (global for simplicity; consider a state store like DynamoDB for production)
var (
	buffer      = make(map[string][]TelemetryData)
	lastValues  = map[string]float64{"FLOWRATE": 0, "PRESSURE": 0, "TEMPERATURE": 0}
	bufferMutex sync.Mutex
)

func Handler(ctx context.Context, kinesisEvent events.KinesisEvent, apiGateway *apigatewaymanagementapi.ApiGatewayManagementApi) error {
	// Retrieve WebSocket connections
	connections, err := websocket.GetActiveConnections()
	if err != nil {
		fmt.Printf("Failed to retrieve connections: %v\n", err)
		return err
	}

	// Process each Kinesis record
	for _, record := range kinesisEvent.Records {
		dataBytes := record.Kinesis.Data
		fmt.Printf("Received Kinesis record: %s\n", dataBytes)

		var telemetryData TelemetryData
		if err := json.Unmarshal(dataBytes, &telemetryData); err != nil {
			fmt.Printf("Cannot read Kinesis telemetry data: %v\n", err)
			continue
		}

		fmt.Printf("Telemetry Data: %+v\n", telemetryData)

		// Buffer into 5-second window
		if err := bufferData(telemetryData); err != nil {
			fmt.Printf("Error buffering data: %v\n", err)
			continue
		}
	}

	// Process completed 5-second buckets
	processedBuckets := processCompletedBuckets()
	for _, bucket := range processedBuckets {
		// Get anomaly score from SageMaker RCF
		features := []float64{
			bucket.FLOWRATE, bucket.PRESSURE, bucket.TEMPERATURE,
			bucket.FlowChangeRate, bucket.PressChangeRate, bucket.TempChangeRate,
		}

		bucket.AnomalyScore = sagemaker.Predict(features) // Assumes SageMaker expects a 6D vector

		// Marshal response
		responseBytes, err := json.Marshal(bucket)
		if err != nil {
			fmt.Printf("Error marshaling response: %v\n", err)
			continue
		}

		// Send to WebSocket clients
		for _, conn := range connections {
			_, err := apiGateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(conn.ConnectionID),
				Data:         responseBytes,
			})
			if err != nil {
				fmt.Printf("Error sending to %s: %v\n", conn.ConnectionID, err)
			}
		}
	}

	return nil
}

// bufferData adds telemetry data to a 5-second bucket
func bufferData(data TelemetryData) error {
	ts, err := time.Parse(time.RFC3339, data.Timestamp)
	if err != nil {
		return fmt.Errorf("invalid timestamp: %v", err)
	}

	// Floor to nearest 5-second boundary
	bucketStart := ts.Truncate(5 * time.Second)
	bucketKey := bucketStart.Format(time.RFC3339)

	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	buffer[bucketKey] = append(buffer[bucketKey], data)
	return nil
}

// processCompletedBuckets processes all buckets older than 5 seconds from now
func processCompletedBuckets() []ProcessedData {
	bufferMutex.Lock()
	defer bufferMutex.Unlock()

	now := time.Now().UTC()
	var processed []ProcessedData

	for bucketKey, data := range buffer {
		bucketStart, err := time.Parse(time.RFC3339, bucketKey)
		if err != nil {
			fmt.Printf("Invalid bucket key %s: %v\n", bucketKey, err)
			continue
		}

		// Process if bucket is complete (older than 5s from now)
		if now.Sub(bucketStart) >= 5*time.Second {
			processedData := processBucket(bucketKey, data)
			if processedData != nil {
				processed = append(processed, *processedData)
			}
			delete(buffer, bucketKey)
		}
	}

	return processed
}

// processBucket aggregates data into a single 5s row with rates
func processBucket(bucketKey string, data []TelemetryData) *ProcessedData {
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
	processed := ProcessedData{
		Timestamp:       bucketKey,
		FLOWRATE:        current["FLOWRATE"],
		PRESSURE:        current["PRESSURE"],
		TEMPERATURE:     current["TEMPERATURE"],
		FlowChangeRate:  (current["FLOWRATE"] - lastValues["FLOWRATE"]) / 5,
		PressChangeRate: (current["PRESSURE"] - lastValues["PRESSURE"]) / 5,
		TempChangeRate:  (current["TEMPERATURE"] - lastValues["TEMPERATURE"]) / 5,
	}

	// Update last values for next bucket
	lastValues["FLOWRATE"] = current["FLOWRATE"]
	lastValues["PRESSURE"] = current["PRESSURE"]
	lastValues["TEMPERATURE"] = current["TEMPERATURE"]

	return &processed
}
