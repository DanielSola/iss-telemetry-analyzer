package dynamo

import (
	"fmt"
	"time"

	"iss-telemetry-analyzer/src/types"
	"iss-telemetry-analyzer/src/utils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func StoreAnomalyScore(newScore float64) types.StoreAnomalyScoreResult {
	client := GetDynamoDBClient()

	result, err := client.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("AnomalyScores"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {S: aws.String("scores")},
		},
	})

	if err != nil {
		return types.StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to fetch existing scores: %w", err),
		}
	}

	var existingScores []map[string]interface{}

	if result.Item != nil && len(result.Item) > 0 {
		// Extract the "scores" attribute from the item
		if scoresAttr, ok := result.Item["scores"]; ok {
			// Unmarshal the "scores" attribute into a []map[string]interface{}
			if err := dynamodbattribute.Unmarshal(scoresAttr, &existingScores); err != nil {
				fmt.Printf("Failed to unmarshal existing scores: %v\n", err)
				return types.StoreAnomalyScoreResult{
					Error: fmt.Errorf("failed to unmarshal existing scores: %w", err),
				}
			}
		} else {
			// Initialize an empty array if the "scores" attribute does not exist
			existingScores = []map[string]interface{}{}
		}
	} else {
		// Initialize an empty array if the table is empty or key does not exist
		existingScores = []map[string]interface{}{}
	}

	// Filter out scores older than 1 hour
	oneHourAgo := time.Now().UTC().Add(-1 * time.Minute)
	var filteredScores []map[string]interface{}
	for _, scoreEntry := range existingScores {
		if timestampStr, ok := scoreEntry["timestamp"].(string); ok {
			timestamp, err := time.Parse(time.RFC3339, timestampStr)
			if err != nil {
				fmt.Printf("Error parsing timestamp: %v\n", err)
				continue
			}
			if timestamp.After(oneHourAgo) {
				filteredScores = append(filteredScores, scoreEntry)
			}
		}
	}

	// Calculate the average and std of filteredScores
	var scoresOnly []float64
	for _, scoreEntry := range filteredScores {
		if score, ok := scoreEntry["anomalyScore"].(float64); ok {
			scoresOnly = append(scoresOnly, score)
		}
	}

	avgAnomalyScore := utils.Average(scoresOnly)
	stdAnomalyScore := utils.StandardDeviation(scoresOnly)

	// Append the new score and timestamp to the front of the filtered array
	newEntry := map[string]interface{}{
		"anomalyScore": newScore,
		"timestamp":    time.Now().UTC().Format(time.RFC3339), // Current timestamp in ISO 8601 format
	}

	filteredScores = append([]map[string]interface{}{newEntry}, filteredScores...)

	// Marshal the updated scores array
	item, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"key":    "scores",
		"scores": filteredScores,
	})

	if err != nil {
		return types.StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to marshal updated scores: %w", err),
		}
	}

	// Update the item in the DynamoDB table
	_, err = client.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("AnomalyScores"),
		Item:      item,
	})

	if err != nil {
		return types.StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to store updated scores in table: %w", err),
		}
	}

	return types.StoreAnomalyScoreResult{
		Average:           avgAnomalyScore,
		StandardDeviation: stdAnomalyScore,
		Error:             nil,
		Score:             newScore,
	}
}
