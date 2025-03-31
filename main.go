package main

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

// API Gateway WebSocket connection table (use DynamoDB in production).
var connections = make(map[string]bool)

// Handles Kinesis events and sends data to all WebSocket connections
func handler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	// Initialize API Gateway Management API
	sess := session.Must(session.NewSession())
	apigateway := apigatewaymanagementapi.New(sess)

	// Iterate over each record in the Kinesis event
	for _, record := range kinesisEvent.Records {
		// Decode the Kinesis data (Base64 encoded)
		data := record.Kinesis.Data
		fmt.Printf("Received Kinesis record: %s\n", string(data))

		// Broadcast the data to all WebSocket connections
		for connectionID := range connections {
			_, err := apigateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(connectionID),
				Data:         data, // Send the raw Kinesis data
			})
			if err != nil {
				fmt.Printf("Error sending message to connection %s: %v\n", connectionID, err)
			}
		}
	}

	return nil
}

// WebSocket connection management handlers
func manageWebSocket(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.RequestContext.RouteKey {
	case "$connect":
		fmt.Println("New connection:", req.RequestContext.ConnectionID)
		connections[req.RequestContext.ConnectionID] = true
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	case "$disconnect":
		fmt.Println("Disconnected:", req.RequestContext.ConnectionID)
		delete(connections, req.RequestContext.ConnectionID)
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	default:
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}, nil
	}
}

func main() {
	// Start the Lambda handler
	lambda.Start(handler)
}
