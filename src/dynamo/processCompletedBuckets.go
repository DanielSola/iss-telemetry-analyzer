package dynamo

import (
	"fmt"
	"iss-telemetry-analyzer/src/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func ProcessCompletedBuckets() []types.ProcessedData {
	now := time.Now().UTC()
	var processed []types.ProcessedData

	client := GetDynamoDBClient()

	// Scan DynamoDB for buckets older than 5 seconds
	result, err := client.Scan(&dynamodb.ScanInput{TableName: aws.String("TelemetryBucket")})

	if err != nil {
		fmt.Printf("Failed to scan DynamoDB: %v\n", err)
		return processed
	}

	for _, item := range result.Items {

		var record struct {
			BucketKey string                `json:"BucketKey"`
			Data      []types.TelemetryData `json:"Data"`
		}

		if err := dynamodbattribute.UnmarshalMap(item, &record); err != nil {
			fmt.Printf("Failed to unmarshal DynamoDB item: %v\nRaw item: %+v\n", err, item)
			continue
		}

		bucketStart, err := time.Parse(time.RFC3339, record.BucketKey)

		if err != nil {
			fmt.Printf("Invalid bucket key %s: %v\n", record.BucketKey, err)
			continue
		}

		// Process if bucket is complete (older than 5s from now)
		if now.Sub(bucketStart) >= 5*time.Second {
			processedData := ProcessBucket(record.BucketKey, record.Data)
			if processedData != nil {
				processed = append(processed, *processedData)
			}

			// Delete processed bucket from DynamoDB
			_, err := client.DeleteItem(&dynamodb.DeleteItemInput{
				TableName: aws.String("TelemetryBucket"),
				Key: map[string]*dynamodb.AttributeValue{
					"BucketKey": {S: aws.String(record.BucketKey)},
				},
			})
			if err != nil {
				fmt.Printf("Failed to delete bucket %s: %v\n", record.BucketKey, err)
			}
		}
	}

	return processed
}
