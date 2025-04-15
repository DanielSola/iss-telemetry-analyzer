package dynamo

import (
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

type TelemetryData struct {
	PK           string  `json:"PK"`
	SK           string  `json:"SK"`
	Score        float64 `json:"score"`
	Pressure     float64 `json:"pressure"`
	Temperature  float64 `json:"temperature"`
	Flowrate     float64 `json:"flowrate"`
	AnomalyLevel string  `json:"anomaly_level"`
	TTL          int64   `json:"ttl"`
}

func FetchHistoricalData() ([]TelemetryData, error) {
	now := time.Now()
	oneHourAgo := now.Add(-1 * time.Hour).UTC().Format(time.RFC3339) // Convert to ISO 8601 format

	client := GetDynamoDBClient()

	partitionKey := "device"

	input := &dynamodb.QueryInput{
		TableName:              aws.String("TelemetryData"),
		KeyConditionExpression: aws.String("PK = :partitionKey AND SK >= :oneHourAgo"),
		ExpressionAttributeValues: map[string]*dynamodb.AttributeValue{
			":partitionKey": {
				S: aws.String(partitionKey), // Partition key value
			},
			":oneHourAgo": {
				S: aws.String(oneHourAgo), // Sort key condition as ISO 8601 string
			},
		},
	}

	output, err := client.Query(input)
	if err != nil {
		return nil, err
	}

	var results []TelemetryData
	err = dynamodbattribute.UnmarshalListOfMaps(output.Items, &results)
	if err != nil {
		return nil, err
	}

	return results, nil
}
