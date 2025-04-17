package dynamo

import (
	"fmt"
	"time"

	"iss-telemetry-analyzer/src/utils"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

func StoreTelemetryData(score, pressure, temperature, flowrate float64, level utils.AnomalyLevel) error {
	client := GetDynamoDBClient()

	now := time.Now()

	partitionKey := "device" // Static key since you're storing one day of data
	sortKey := now.Format(time.RFC3339)
	ttl := now.Add(24 * time.Hour).Unix()

	input := &dynamodb.PutItemInput{
		TableName: aws.String("TelemetryData"),
		Item: map[string]*dynamodb.AttributeValue{
			"PK": {
				S: aws.String(partitionKey),
			},
			"SK": {
				S: aws.String(sortKey),
			},
			"score": {
				N: aws.String(fmt.Sprintf("%.4f", score)),
			},
			"pressure": {
				N: aws.String(fmt.Sprintf("%.4f", pressure)),
			},
			"temperature": {
				N: aws.String(fmt.Sprintf("%.4f", temperature)),
			},
			"flowrate": {
				N: aws.String(fmt.Sprintf("%.4f", flowrate)),
			},
			"anomaly_level": {
				S: aws.String(level.String()), // Convert AnomalyLevel to string
			},
			"ttl": {
				N: aws.String(fmt.Sprintf("%d", ttl)),
			},
		},
	}

	fmt.Println("StoreTelemetryData", input)

	output, err := client.PutItem(input)

	if err != nil {
		return fmt.Errorf("failed to put item: %w", err)
	}

	fmt.Println("Stored!", output)

	return nil
}
