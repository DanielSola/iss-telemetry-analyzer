package websocket

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

var (
	sess      = session.Must(session.NewSession())
	dynamoDB  = dynamodb.New(sess)
	tableName = "WebSocketConnections"
)

// Fetch all active WebSocket connections from DynamoDB
func GetActiveConnections() ([]WebSocketConnection, error) {
	input := &dynamodb.ScanInput{TableName: aws.String(tableName)}
	result, err := dynamoDB.Scan(input)
	if err != nil {
		return nil, err
	}

	var connections []WebSocketConnection
	err = dynamodbattribute.UnmarshalListOfMaps(result.Items, &connections)
	return connections, err
}
