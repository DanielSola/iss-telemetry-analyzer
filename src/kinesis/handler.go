package kinesis

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/websocket"
	"math/rand"
	"strconv"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

type TelemetryData struct {
	Name      string `json:"name"`      // Name of the telemetry data
	Value     string `json:"value"`     // Value of the telemetry data
	Timestamp string `json:"timestamp"` // Timestamp of the telemetry data
}

type WebSocketResponse struct {
	TelemetryData TelemetryData `json:"telemetryData"` // Original telemetry data
	AnomalyScore  float64       `json:"anomalyScore"`  // Anomaly score
}

func Handler(ctx context.Context, kinesisEvent events.KinesisEvent, apiGateway *apigatewaymanagementapi.ApiGatewayManagementApi) error {
	// Retrieve all active WebSocket connections from DynamoDB
	connections, err := websocket.GetActiveConnections()
	if err != nil {
		fmt.Printf("Failed to retrieve connections: %v\n", err)
		return err
	}

	// Iterate over each record in the Kinesis event
	for _, record := range kinesisEvent.Records {
		// Decode the Kinesis data
		dataBytes := record.Kinesis.Data

		fmt.Printf("Received Kinesis record: %s\n", dataBytes)

		var telemetryData TelemetryData

		err := json.Unmarshal(dataBytes, &telemetryData)

		if err != nil {
			fmt.Print(err)
			fmt.Println("Cannot read kinesis telemetry data")
		}

		fmt.Printf("Telemetry Data: %+v\n", telemetryData)

		// Convert Value to float64
		value, err := strconv.ParseFloat(telemetryData.Value, 64)

		if err != nil {
			fmt.Printf("Error converting value to float64: %v\n", err)
			continue
		} else {
			fmt.Println("Received value: ", value)
		}

		anomalyScore := rand.Float64()

		response := WebSocketResponse{
			TelemetryData: telemetryData,
			AnomalyScore:  anomalyScore,
		}

		responseBytes, err := json.Marshal(response)

		if err != nil {
			fmt.Printf("Error marshaling WebSocket response: %v\n", err)
		}

		// Send data to all active WebSocket connections
		for _, connection := range connections {
			_, err := apiGateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(connection.ConnectionID),
				Data:         responseBytes,
			})

			if err != nil {
				fmt.Printf("Error sending message to connection %s: %v\n", connection.ConnectionID, err)
			}
		}
	}

	return nil
}
