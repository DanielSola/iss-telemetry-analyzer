package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
)

// Define request and response structure
type Request struct {
	Name string `json:"name"`
}

type Response struct {
	Message string `json:"message"`
}

// Lambda function handler
func handler(ctx context.Context, req Request) (Response, error) {
	message := fmt.Sprintf("Hello, %s! Your Go Lambda is working locally!", req.Name)
	return Response{Message: message}, nil
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "local" {
		// Local testing mode
		testEvent := Request{Name: "LocalUser"}
		response, _ := handler(context.Background(), testEvent)
		jsonResponse, _ := json.Marshal(response)
		fmt.Println(string(jsonResponse))
	} else {
		// Run on AWS Lambda
		lambda.Start(handler)
	}
}
