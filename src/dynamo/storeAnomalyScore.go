package dynamo

import (
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func StoreAnomalyScore(dynamoDBClient *dynamodb.DynamoDB, newScore float64) error {
	// Retrieve scores array
	result, err := dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("AnomalyScores"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {S: aws.String("scores")},
		},
	})

	if err != nil {
		return fmt.Errorf("failed to fetch existing scores: %w", err)
	}

	var existingScores []float64

	if result.Item != nil && len(result.Item) > 0 {
		// Unmarshal existing data if the item exists
		if err := dynamodbattribute.UnmarshalMap(result.Item, &existingScores); err != nil {
			fmt.Printf("Failed to unmarshal existing scores: %v\n", err)
			return fmt.Errorf("failed to unmarshal existing scores: %w", err)
		}
	} else {
		// Initialize an empty array if the table is empty or key does not exist
		existingScores = []float64{}
	}

	// Append the new score to the existing array
	existingScores = append(existingScores, newScore)

	// Marshal the updated scores array
	item, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"key":    "scores",
		"scores": existingScores,
	})

	if err != nil {
		return fmt.Errorf("failed to marshal updated scores: %w", err)
	}

	// Update the item in the DynamoDB table
	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("AnomalyScores"),
		Item:      item,
	})

	if err != nil {
		return fmt.Errorf("failed to store updated scores in table: %w", err)
	}

	return nil
}
