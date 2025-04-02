package kinesis

import (
	"context"
	"encoding/base64"
	"fmt"
	"iss-telemetry-analyzer/src/websocket"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

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
		dataBytes, err := base64.StdEncoding.DecodeString(string(record.Kinesis.Data))

		if err != nil {
			fmt.Printf("Failed to decode Kinesis data: %v\n", err)
			continue
		}

		fmt.Printf("Received Kinesis record: %s\n", dataBytes)

		// Send data to all active WebSocket connections
		for _, connection := range connections {
			_, err := apiGateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(connection.ConnectionID),
				Data:         dataBytes,
			})
			if err != nil {
				fmt.Printf("Error sending message to connection %s: %v\n", connection.ConnectionID, err)
			}
		}
	}

	return nil
}
