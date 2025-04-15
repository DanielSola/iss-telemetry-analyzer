package kinesis

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func DetectEventType(event json.RawMessage) (string, error) {
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

	// Try parsing as an API Gateway HTTP request (v2)
	var httpEvent events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(event, &httpEvent); err == nil {
		if httpEvent.RequestContext.HTTP.Method != "" {
			return "http", nil
		}
	}

	return "", fmt.Errorf("unknown event type")
}
