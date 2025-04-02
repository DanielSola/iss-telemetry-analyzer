package websocket

import (
	"context"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// WebSocketConnection represents a connection stored in DynamoDB
type WebSocketConnection struct {
	ConnectionID string `json:"connectionId"`
}

// WebSocket connection management handlers
func Manage(ctx context.Context, req events.APIGatewayWebsocketProxyRequest, apiGateway *apigatewaymanagementapi.ApiGatewayManagementApi) (events.APIGatewayProxyResponse, error) {
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
