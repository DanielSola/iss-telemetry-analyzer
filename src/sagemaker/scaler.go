package sagemaker

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

// RobustScalerParams holds the parameters for the robust scaler
type RobustScalerParams struct {
	Medians []float64 `json:"center"` // Updated field name to match JSON
	IQRs    []float64 `json:"scale"`  // Updated field name to match JSON
}

// LoadRobustScalerParams loads the robust scaler parameters from S3
func LoadRobustScalerParams() (*RobustScalerParams, error) {
	// Create an S3 client
	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("eu-west-1"),
	}))
	svc := s3.New(sess)

	// Download the parameters
	bucketName := os.Getenv("S3_BUCKET_NAME")
	if bucketName == "" {
		return nil, fmt.Errorf("S3_BUCKET_NAME environment variable is not set")
	}

	result, err := svc.GetObject(&s3.GetObjectInput{
		Bucket: aws.String(bucketName),
		Key:    aws.String("models/scaler_params.json"),
	})

	if err != nil {
		return nil, fmt.Errorf("failed to download scaler parameters: %w", err)
	}
	defer result.Body.Close()

	// Read the file content
	content, err := ioutil.ReadAll(result.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read scaler parameters: %w", err)
	}

	// Parse the JSON
	var params RobustScalerParams
	err = json.Unmarshal(content, &params)
	if err != nil {
		return nil, fmt.Errorf("failed to parse scaler parameters: %w", err)
	}

	return &params, nil
}

// RobustScale applies robust scaling to the features
func RobustScale(features []float64, params *RobustScalerParams) []float64 {
	if len(features) != len(params.Medians) {
		// Handle error: feature vector length doesn't match parameters
		return features
	}

	scaledFeatures := make([]float64, len(features))

	for i, val := range features {
		if params.IQRs[i] != 0 {
			scaledFeatures[i] = (val - params.Medians[i]) / params.IQRs[i]
		} else {
			scaledFeatures[i] = val - params.Medians[i] // If IQR is 0, just center the feature
		}
	}

	return scaledFeatures
}
