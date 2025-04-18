package dynamo

import (
	"fmt"
	"iss-telemetry-analyzer/src/types"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func BufferData(data types.TelemetryData) error {
	client := GetDynamoDBClient()

	ts, err := time.Parse(time.RFC3339, data.Timestamp)

	if err != nil {
		return fmt.Errorf("invalid timestamp: %v", err)
	}

	// Floor to nearest 5-second boundary
	bucketStart := ts.Truncate(5 * time.Second)
	bucketKey := bucketStart.Format(time.RFC3339)

	// Retrieve existing bucket data
	result, err := client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("TelemetryBucket"),
		Key: map[string]*dynamodb.AttributeValue{
			"BucketKey": {S: aws.String(bucketKey)},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to retrieve bucket: %v", err)
	}

	var existingBucket types.DynamoData

	if result.Item != nil {
		// Unmarshal existing data if the bucket exists
		if err := dynamodbattribute.UnmarshalMap(result.Item, &existingBucket); err != nil {
			fmt.Printf("Failed to unmarshal existing bucket data: %v\n", err)
			return fmt.Errorf("failed to unmarshal existing bucket data: %v", err)
		}
	}

	// Append new data to the existing slice
	existingBucket.Data = append(existingBucket.Data, data)

	if existingBucket.BucketKey == nil {
		existingBucket.BucketKey = aws.String(bucketKey) // Set it to the current bucket key
	}

	// Marshal updated data to DynamoDB format
	item, err := dynamodbattribute.MarshalMap(existingBucket)

	if err != nil {
		return fmt.Errorf("failed to marshal data: %v", err)
	}

	// Save updated bucket to DynamoDB
	_, err = client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("TelemetryBucket"),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to save data to DynamoDB: %v", err)
	}

	return nil
}
