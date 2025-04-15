package websocket

import (
	"fmt"
	"os"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

var (
	apiGatewayClient *apigatewaymanagementapi.ApiGatewayManagementApi
	once             sync.Once
	apiURL           = os.Getenv("API_GATEWAY_URL")
)

func GetApiGWClient() *apigatewaymanagementapi.ApiGatewayManagementApi {
	if apiURL == "" {
		fmt.Println("Error: API_GATEWAY_URL environment variable not set")
		os.Exit(1)
	}

	once.Do(func() {
		sess := session.Must(session.NewSession())
		apiGatewayClient = apigatewaymanagementapi.New(sess, aws.NewConfig().WithEndpoint(apiURL))
	})

	return apiGatewayClient
}
