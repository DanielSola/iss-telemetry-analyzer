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

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
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

		var telemetryData types.TelemetryData

		if err := json.Unmarshal(dataBytes, &telemetryData); err != nil {
			fmt.Printf("Cannot read Kinesis telemetry data: %v\n", err)
			continue
		}

		fmt.Printf("Telemetry Data: %+v\n", telemetryData)

		// Buffer into 5-second window
		if err := dynamo.BufferData(telemetryData); err != nil {
			fmt.Printf("Error buffering data: %v\n", err)
			continue
		}
	}

	// Process completed 5-second buckets
	processedBuckets := dynamo.ProcessCompletedBuckets()

	for _, bucket := range processedBuckets {
		// Get anomaly score from SageMaker RCF
		features := []float64{
			bucket.FLOWRATE, bucket.PRESSURE, bucket.TEMPERATURE,
			bucket.FlowChangeRate, bucket.PressChangeRate, bucket.TempChangeRate,
		}

		var anomalyScore = sagemaker.Predict(features)

		bucket.AnomalyScore = anomalyScore

		result := dynamo.StoreAnomalyScore(anomalyScore)

		anomalyLevel := utils.ComputeAnomalyLevel(result.Score, result.StandardDeviation, result.Average)

		bucket.AnomalyLevel = string(anomalyLevel)

		if result.Error != nil {
			fmt.Printf("Error storing anomaly score: %v\n", result.Error)
		} else {
			fmt.Printf("Score: %f, Average: %f, Standard Deviation: %f, Anomaly Level: %s\n", result.Score, result.Average, result.StandardDeviation, anomalyLevel)
		}

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
