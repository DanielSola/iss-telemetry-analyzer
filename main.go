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

var (
	sagemakerEndpointURL = os.Getenv("SAGEMAKER_ENDPOINT_NAME") // TODO: Leer el endpoint de aqui y hacer predicciones
)

// Determine which handler to start based on event type
func main() {
	lambda.Start(func(ctx context.Context, event json.RawMessage) (interface{}, error) {
		eventType, err := kinesis.DetectEventType(event)

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
		default:
			// Make sure you handle unknown events properly.
			return nil, fmt.Errorf("unknown event type: %s", eventType)
		}
	})
}
