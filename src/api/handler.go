package api

import (
	"context"
	"encoding/json"
	"iss-telemetry-analyzer/src/dynamo"

	"github.com/aws/aws-lambda-go/events"
)

func HandleHTTP(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {

	switch req.RouteKey {
	case "GET /history":
		// Use query parameters like ?range=1h
		data, err := dynamo.FetchHistoricalData()

		if err != nil {
			return events.APIGatewayV2HTTPResponse{StatusCode: 500, Body: err.Error()}, nil
		}

		body, _ := json.Marshal(data)

		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers:    map[string]string{"Content-Type": "application/json"},
			Body:       string(body),
		}, nil

	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Body:       "Not Found",
		}, nil
	}
}
