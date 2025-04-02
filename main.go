package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/kinesis"
	"iss-telemetry-analyzer/src/websocket"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

type Record struct {
	EventSource    string
	EventSourceArn string
	AWSRegion      string
	S3             events.S3Entity
	SQS            events.SQSMessage
	SNS            events.SNSEntity
}

// Event incoming event
type Event struct {
	Records []Record
}

func detectEventType(event json.RawMessage) (string, error) {
	// Try parsing as a Kinesis event
	var kinesisEvent events.KinesisEvent
	if err := json.Unmarshal(event, &kinesisEvent); err == nil {
		if len(kinesisEvent.Records) > 0 && kinesisEvent.Records[0].EventSource == "aws:kinesis" {
			return "kinesis", nil
		}
	}

	// Try parsing as an API Gateway WebSocket event
	var websocketEvent events.APIGatewayWebsocketProxyRequest
	if err := json.Unmarshal(event, &websocketEvent); err == nil {
		if websocketEvent.RequestContext.EventType != "" {
			return "websocket", nil
		}
	}

	return "", fmt.Errorf("unknown event type")
}

var (
	sess       = session.Must(session.NewSession())
	apiURL     = os.Getenv("API_GATEWAY_URL") // Get API Gateway URL from environment variable
	apiGateway *apigatewaymanagementapi.ApiGatewayManagementApi
)

// Initialize API Gateway Management API dynamically
func init() {
	if apiURL == "" {
		fmt.Println("Error: API_GATEWAY_URL environment variable not set")
		os.Exit(1)
	}
	apiGateway = apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(apiURL))
}

// Determine which handler to start based on event type
func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		eventType, err := detectEventType(event)

		if err != nil {
			fmt.Println("Error detecting event type:", err)
			return nil, err
		}

		// Event handling based on the event type
		switch eventType {
		case "kinesis":
			// Unmarshal the event into KinesisEvent
			var kinesisEvent events.KinesisEvent
			err := json.Unmarshal(event, &kinesisEvent)

			if err != nil {
				fmt.Printf("Error unmarshalling Kinesis event: %v\n", err)
				return nil, err
			}

			// You likely want to return the result or a message instead of nil.
			return nil, kinesis.Handler(ctx, kinesisEvent, apiGateway)
		case "websocket":
			// Unmarshal the event into KinesisEvent
			var websocketEvent events.APIGatewayWebsocketProxyRequest
			err := json.Unmarshal(event, &websocketEvent)

			if err != nil {
				fmt.Printf("Error unmarshalling Kinesis event: %v\n", err)
				return nil, err
			}

			return websocket.Manage(ctx, websocketEvent, apiGateway)
		default:
			// Make sure you handle unknown events properly.
			return nil, fmt.Errorf("unknown event type: %s", eventType)
		}
	})
}
