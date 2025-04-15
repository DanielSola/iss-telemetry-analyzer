package api

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
)

func fetchHistoricalData(timeRange string) ([]map[string]interface{}, error) {
	// TODO: Replace this with a real DB query
	return []map[string]interface{}{
		{"timestamp": "2025-04-15T12:00:00Z", "value": 42.0},
		{"timestamp": "2025-04-15T12:01:00Z", "value": 43.1},
	}, nil
}

func HandleHTTP(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	fmt.Println("req.RouteKey:", req.RouteKey)
	fmt.Println("req.RawPath:", req.RawPath)
	fmt.Println("req.headers:", req.Headers)
	fmt.Println("req.reqContest:", req.RequestContext)
	fmt.Println("req.RouteKey:", req.RequestContext.RouteKey)
	fmt.Println("req.method:", req.RequestContext.HTTP.Method)
	fmt.Println("req.http.path:", req.RequestContext.HTTP.Path)

	switch req.RequestContext.HTTP.Path {
	case "/history":
		// Use query parameters like ?range=1h
		timeRange := req.QueryStringParameters["range"]
		data, err := fetchHistoricalData(timeRange)
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
