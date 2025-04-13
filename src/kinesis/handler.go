package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/dynamo"
	"iss-telemetry-analyzer/src/sagemaker"
	"iss-telemetry-analyzer/src/websocket"
	"strconv"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type DynamoData struct {
	BucketKey *string         `dynamodbav:"BucketKey"` // Correct tag for DynamoDB
	Data      []TelemetryData `dynamodbav:"Data"`      // Correct tag for DynamoDB
}

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

var (
	dynamoDBClient    *dynamodb.DynamoDB
	dynamoDBTableName = "TelemetryData"
	lastValues        = map[string]float64{"FLOWRATE": 0, "PRESSURE": 0, "TEMPERATURE": 0}
	chunkSizeSeconds  = 5
)

func init() {
	// Initialize DynamoDB client
	dynamoDBClient = dynamodb.New(session.Must(session.NewSession()))
}

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

		var anomalyScore = sagemaker.Predict(features)

		bucket.AnomalyScore = anomalyScore

		fmt.Println("Store anomaly score", anomalyScore)
		dynamo.StoreAnomalyScore(dynamoDBClient, anomalyScore)

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

// bufferData adds telemetry data to a 5-second bucket in DynamoDB
func bufferData(data TelemetryData) error {
	ts, err := time.Parse(time.RFC3339, data.Timestamp)

	if err != nil {
		return fmt.Errorf("invalid timestamp: %v", err)
	}

	// Floor to nearest 5-second boundary
	bucketStart := ts.Truncate(5 * time.Second)
	bucketKey := bucketStart.Format(time.RFC3339)

	// Retrieve existing bucket data
	result, err := dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String(dynamoDBTableName),
		Key: map[string]*dynamodb.AttributeValue{
			"BucketKey": {S: aws.String(bucketKey)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to retrieve bucket: %v", err)
	}

	var existingBucket DynamoData

	if result.Item != nil {
		// Unmarshal existing data if the bucket exists
		if err := dynamodbattribute.UnmarshalMap(result.Item, &existingBucket); err != nil {
			fmt.Printf("Failed to unmarshal existing bucket data: %v\n", err)
			return fmt.Errorf("failed to unmarshal existing bucket data: %v", err)
		}
	}

	// Append new data to the existing slice
	existingBucket.Data = append(existingBucket.Data, data)

	if existingBucket.BucketKey == nil {
		existingBucket.BucketKey = aws.String(bucketKey) // Set it to the current bucket key
	}

	// Marshal updated data to DynamoDB format
	item, err := dynamodbattribute.MarshalMap(existingBucket)

	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Save updated bucket to DynamoDB
	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(dynamoDBTableName),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save data to DynamoDB: %v", err)
	}

	return nil
}

// processCompletedBuckets processes all buckets older than 5 seconds from now
func processCompletedBuckets() []ProcessedData {
	now := time.Now().UTC()
	var processed []ProcessedData

	// Scan DynamoDB for buckets older than 5 seconds
	result, err := dynamoDBClient.Scan(&dynamodb.ScanInput{TableName: aws.String(dynamoDBTableName)})

	if err != nil {
		fmt.Printf("Failed to scan DynamoDB: %v\n", err)
		return processed
	}

	for _, item := range result.Items {

		var record struct {
			BucketKey string          `json:"BucketKey"`
			Data      []TelemetryData `json:"Data"`
		}

		if err := dynamodbattribute.UnmarshalMap(item, &record); err != nil {
			fmt.Printf("Failed to unmarshal DynamoDB item: %v\nRaw item: %+v\n", err, item)
			continue
		}

		bucketStart, err := time.Parse(time.RFC3339, record.BucketKey)

		if err != nil {
			fmt.Printf("Invalid bucket key %s: %v\n", record.BucketKey, err)
			continue
		}

		// Process if bucket is complete (older than 5s from now)
		if now.Sub(bucketStart) >= 5*time.Second {
			processedData := processBucket(record.BucketKey, record.Data)
			if processedData != nil {
				processed = append(processed, *processedData)
			}

			// Delete processed bucket from DynamoDB
			_, err := dynamoDBClient.DeleteItem(&dynamodb.DeleteItemInput{
				TableName: aws.String(dynamoDBTableName),
				Key: map[string]*dynamodb.AttributeValue{
					"BucketKey": {S: aws.String(record.BucketKey)},
				},
			})
			if err != nil {
				fmt.Printf("Failed to delete bucket %s: %v\n", record.BucketKey, err)
			}
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
