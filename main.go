package main

import (
	"context"
	"fmt"
	"math/rand"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

// API Gateway WebSocket connection table (use DynamoDB in production).
var connections = make(map[string]bool)

// Handle WebSocket events
func handler(ctx context.Context, req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	switch req.RequestContext.RouteKey {
	case "$connect":
		fmt.Println("New connection:", req.RequestContext.ConnectionID)
		connections[req.RequestContext.ConnectionID] = true
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	case "$disconnect":
		fmt.Println("Disconnected:", req.RequestContext.ConnectionID)
		delete(connections, req.RequestContext.ConnectionID)
		return events.APIGatewayProxyResponse{StatusCode: 200}, nil

	case "sendmessage": // Custom WebSocket route
		return handleSendMessage(req)

	default:
		return events.APIGatewayProxyResponse{StatusCode: 400, Body: "Invalid request"}, nil
	}
}

// Handles sending random telemetry data to all connected clients
func handleSendMessage(req events.APIGatewayWebsocketProxyRequest) (events.APIGatewayProxyResponse, error) {
	sess := session.Must(session.NewSession())
	apigateway := apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(req.RequestContext.DomainName+"/"+req.RequestContext.Stage))

	randomNumber := rand.Intn(100)
	message := fmt.Sprintf("Random Telemetry: %d", randomNumber)

	for connectionID := range connections {
		_, err := apigateway.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{
			ConnectionId: aws.String(connectionID),
			Data:         []byte(message),
		})
		if err != nil {
			fmt.Println("Error sending message:", err)
		}
	}

	return events.APIGatewayProxyResponse{StatusCode: 200}, nil
}

func main() {
	lambda.Start(handler)
}
