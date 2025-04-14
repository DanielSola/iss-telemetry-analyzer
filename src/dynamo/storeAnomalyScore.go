package dynamo

import (
	"fmt"
	"math"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

func average(xs []float64) float64 {
	total := 0.0
	for _, v := range xs {
		total += v
	}
	return total / float64(len(xs))
}

func standardDeviation(xs []float64) float64 {
	if len(xs) == 0 {
		return 0.0
	}

	mean := average(xs)
	var varianceSum float64

	for _, v := range xs {
		varianceSum += math.Pow(v-mean, 2)
	}

	variance := varianceSum / float64(len(xs))
	return math.Sqrt(variance)
}

type StoreAnomalyScoreResult struct {
	Score             float64
	Average           float64
	StandardDeviation float64
	Error             error
}

func StoreAnomalyScore(dynamoDBClient *dynamodb.DynamoDB, newScore float64) StoreAnomalyScoreResult {
	// Retrieve scores array
	result, err := dynamoDBClient.GetItem(&dynamodb.GetItemInput{
		TableName: aws.String("AnomalyScores"),
		Key: map[string]*dynamodb.AttributeValue{
			"key": {S: aws.String("scores")},
		},
	})

	fmt.Println("Error fetch", err)

	if err != nil {
		return StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to fetch existing scores: %w", err),
		}
	}

	var existingScores []float64

	if result.Item != nil && len(result.Item) > 0 {
		// Extract the "scores" attribute from the item
		if scoresAttr, ok := result.Item["scores"]; ok {
			// Unmarshal the "scores" attribute into a []float64
			if err := dynamodbattribute.Unmarshal(scoresAttr, &existingScores); err != nil {
				fmt.Printf("Failed to unmarshal existing scores: %v\n", err)
				return StoreAnomalyScoreResult{
					Error: fmt.Errorf("failed to unmarshal existing scores: %w", err),
				}
			}
		} else {
			// Initialize an empty array if the "scores" attribute does not exist
			existingScores = []float64{}
		}
	} else {
		// Initialize an empty array if the table is empty or key does not exist
		existingScores = []float64{}
	}

	// Calculate the average and std of existingScores
	avgAnomalyScore := average(existingScores)
	stdAnomalyScore := standardDeviation((existingScores))

	// Append the new score to the front of the existing array
	existingScores = append([]float64{newScore}, existingScores...)

	// Limit the array to the most recent 24 values. 24 x 5 = 120 seconds of memory = 2 mins
	if len(existingScores) > 25 {
		existingScores = existingScores[:25]
	}

	// Marshal the updated scores array
	item, err := dynamodbattribute.MarshalMap(map[string]interface{}{
		"key":    "scores",
		"scores": existingScores,
	})

	if err != nil {
		return StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to marshal updated scores: %w", err),
		}
	}

	// Update the item in the DynamoDB table
	_, err = dynamoDBClient.PutItem(&dynamodb.PutItemInput{
		TableName: aws.String("AnomalyScores"),
		Item:      item,
	})

	if err != nil {
		return StoreAnomalyScoreResult{
			Error: fmt.Errorf("failed to store updated scores in table: %w", err),
		}
	}

	return StoreAnomalyScoreResult{
		Average:           avgAnomalyScore,
		StandardDeviation: stdAnomalyScore,
		Error:             nil,
		Score:             newScore,
	}
}
