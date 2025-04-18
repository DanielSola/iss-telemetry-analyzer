package websocket

import (
	"context"
	"iss-telemetry-analyzer/src/dynamo"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// WebSocketConnection represents a connection stored in DynamoDB
type WebSocketConnection struct {
	ConnectionID string `json:"connectionId"`
}

// WebSocket connection management handlers
func Manage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.RequestContext.RouteKey {
	case "$connect":
		err := storeConnection(req.RequestContext.ConnectionID)
		if err != nil {
			return events.APIGatewayProxyResponse{StatusCode: 500, Body: "Failed to store connection"}, err
		}
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	case "$disconnect":
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
	client := dynamo.GetDynamoDBClient()

	item, _ := dynamodbattribute.MarshalMap(WebSocketConnection{ConnectionID: connectionID})
	_, err := client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String(tableName),
		Item:      item,
	})
	return err
}

// Remove a WebSocket connection from DynamoDB
func deleteConnection(connectionID string) error {
	client := dynamo.GetDynamoDBClient()

	_, err := client.DeleteItem(&dynamodb.DeleteItemInput{
		TableName: aws.String(tableName),
		Key:       map[string]*dynamodb.AttributeValue{"connectionId": {S: aws.String(connectionID)}},
	})

	return err
}
