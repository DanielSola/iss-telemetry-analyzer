package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	sess       = session.Must(session.NewSession())
	dynamoDB   = dynamodb.New(sess)
	apiURL     = os.Getenv("API_GATEWAY_URL") // Get API Gateway URL from environment variable
	apiGateway *apigatewaymanagementapi.ApiGatewayManagementApi
	tableName  = "WebSocketConnections"
)

// Initialize API Gateway Management API dynamically
func init() {
	if apiURL == "" {
		fmt.Println("Error: API_GATEWAY_URL environment variable not set")
		os.Exit(1)
	}
	apiGateway = apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(apiURL))
}

// WebSocketConnection represents a connection stored in DynamoDB
type WebSocketConnection struct {
	ConnectionID string `json:"connectionId"`
}

// Handles Kinesis events and sends data to all WebSocket connections
func kinesisHandler(ctx context.Context, kinesisEvent events.KinesisEvent) error {
	// Retrieve all active WebSocket connections from DynamoDB
	connections, err := getActiveConnections()
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

// Fetch all active WebSocket connections from DynamoDB
func getActiveConnections() ([]WebSocketConnection, error) {
	input := &dynamodb.ScanInput{TableName: aws.String(tableName)}
	result, err := dynamoDB.Scan(input)
	if err != nil {
		return nil, err
	}

	var connections []WebSocketConnection
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &connections)
	return connections, err
}

// WebSocket connection management handlers
func manageWebSocket(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.RequestContext.RouteKey {
	case "$connect":
		fmt.Println("New connection:", req.RequestContext.ConnectionID)
		err := storeConnection(req.RequestContext.ConnectionID)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to store connection"}, err
		}
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	case "$disconnect":
		fmt.Println("Disconnected:", req.RequestContext.ConnectionID)
		err := deleteConnection(req.RequestContext.ConnectionID)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to delete connection"}, err
		}
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	default:
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}, nil
	}
}

// Store a new WebSocket connection in DynamoDB
func storeConnection(connectionID string) error {
	item, _ := dynamodbattribute.MarshalMap(WebSocketConnection{ConnectionID: connectionID})
	_, err := dynamoDB.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// Remove a WebSocket connection from DynamoDB
func deleteConnection(connectionID string) error {
	_, err := dynamoDB.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       map[string]*dynamodb.AttributeValue{"connectionId": {S: aws.String(connectionID)}},
	})
	return err
}

// Determine which handler to start based on event type
func main() {
	lambda.Start(func(ctx context.Context, event interface{}) (interface{}, error) {
		eventJSON, err := json.MarshalIndent(event, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling event:", err)
		} else {
			fmt.Println("Event:", string(eventJSON))
		}

		switch e := event.(type) {
		case events.KinesisEvent:
			return nil, kinesisHandler(ctx, e)
		case events.APIGatewayWebsocketProxyRequest:
			return manageWebSocket(ctx, e)
		default:
			return nil, fmt.Errorf("unknown event type")
		}
	})
}
