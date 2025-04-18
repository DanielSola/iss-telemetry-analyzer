package main

import (
	"context"
	"encoding/json"
	"fmt"
	"iss-telemetry-analyzer/src/api"
	"iss-telemetry-analyzer/src/kinesis"
	"iss-telemetry-analyzer/src/websocket"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

// Determine which handler to start based on event type
func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		eventType, err := kinesis.DetectEventType(event)

		if err != nil {
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
			return nil, kinesis.Handler(ctx, kinesisEvent)

		case "websocket":
			// Unmarshal the event into KinesisEvent
			var websocketEvent events.APIGatewayWebsocketProxyRequest
			err := json.Unmarshal(event, &websocketEvent)

			if err != nil {
				fmt.Printf("Error unmarshalling Kinesis event: %v\n", err)
				return nil, err
			}

			return websocket.Manage(ctx, websocketEvent)
		case "http":
			var httpEvent events.APIGatewayV2HTTPRequest
			if err := json.Unmarshal(event, &httpEvent); err != nil {
				fmt.Printf("Error unmarshalling HTTP event: %v\n", err)
				return nil, err
			}

			return api.HandleHTTP(ctx, httpEvent)

		default:
			// Make sure you handle unknown events properly.
			return nil, fmt.Errorf("unknown event type: %s", eventType)
		}
	})
}
