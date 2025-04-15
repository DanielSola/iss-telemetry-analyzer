package websocket

import (
	"iss-telemetry-analyzer/src/dynamo"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var tableName = "WebSocketConnections"

// Fetch all active WebSocket connections from DynamoDB
func GetActiveConnections() ([]WebSocketConnection, error) {
	client := dynamo.GetDynamoDBClient()

	input := &dynamodb.ScanInput{TableName: aws.String(tableName)}
	result, err := client.Scan(input)

	if err != nil {
		return nil, err
	}

	var connections []WebSocketConnection
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &connections)
	return connections, err
}
