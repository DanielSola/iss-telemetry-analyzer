package sagemaker

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/sagemakerruntime"
)

type SageMakerResponse struct {
	Scores []struct {
		Score float64 `json:"score"`
	} `json:"scores"`
}

func Predict(value float64) float64 {
	// Read endpoint name from environment variable
	endpointName := os.Getenv("SAGEMAKER_ENDPOINT_NAME")

	if endpointName == "" {
		log.Fatal("Could not get sagemaker endpoint")
	}

	// Load AWS config with eu-west-1 region
	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("eu-west-1"))

	if err != nil {
		log.Fatalf("unable to load AWS config: %v", err)
	}

	// Create SageMaker Runtime client
	client := sagemakerruntime.NewFromConfig(cfg)

	// Prepare payload (Modify based on model input format)
	payload := map[string]interface{}{
		"instances": []map[string]any{
			{"features": []float64{value}}, // Example input
		},
	}

	payloadBytes, _ := json.Marshal(payload)

	// Invoke SageMaker endpoint
	output, err := client.InvokeEndpoint(context.TODO(), &sagemakerruntime.InvokeEndpointInput{
		EndpointName: &endpointName,
		Body:         payloadBytes,
		ContentType:  aws.String("application/json"),
		Accept:       aws.String("application/json"),
	})

	if err != nil {
		log.Fatalf("failed to invoke endpoint: %v", err)
	}

	var response SageMakerResponse

	err = json.Unmarshal(output.Body, &response)

	if err != nil {
		log.Fatalf("failed to parse response: %v", err)
	}

	score := response.Scores[0].Score
	fmt.Println("-----------------")
	fmt.Println("Value: ", value)
	fmt.Println("Anomaly Score: ", score)

	return score
}
