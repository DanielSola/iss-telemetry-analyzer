package api

import (
	"context"
	"encoding/json"
	"iss-telemetry-analyzer/src/dynamo"

	"github.com/aws/aws-lambda-go/events"
)

func HandleHTTP(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	corsHeaders := map[string]string{
		"Content-Type":                 "application/json",
		"Access-Control-Allow-Origin":  "*", // Allow all origins (or specify your domain)
		"Access-Control-Allow-Methods": "GET, POST, OPTIONS",
		"Access-Control-Allow-Headers": "Content-Type",
	}

	switch req.RouteKey {
	case "GET /history":
		data, err := dynamo.FetchHistoricalData()
		if err != nil {
			return events.APIGatewayV2HTTPResponse{
				StatusCode: 500,
				Headers:    corsHeaders,
				Body:       err.Error(),
			}, nil
		}

		body, _ := json.Marshal(data)
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 200,
			Headers:    corsHeaders,
			Body:       string(body),
		}, nil

	default:
		return events.APIGatewayV2HTTPResponse{
			StatusCode: 404,
			Headers:    corsHeaders,
			Body:       "Not Found",
		}, nil
	}
}
