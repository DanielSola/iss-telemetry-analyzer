package dynamo

import (
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

var (
	clientInstance *dynamodb.DynamoDB
	once           sync.Once
)

func GetDynamoDBClient() *dynamodb.DynamoDB {
	once.Do(func() {
		sess := session.Must(session.NewSession(&aws.Config{Region: aws.String("eu-west-1")})) // Replace with your region

		clientInstance = dynamodb.New(sess)
	})

	return clientInstance
}
