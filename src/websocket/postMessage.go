package websocket

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
)

func PostMessage(responseBytes []byte) error {
	// Send to WebSocket clients
	connections, err := GetActiveConnections()

	if err != nil {
		fmt.Printf("Failed to retrieve connections: %v\n", err)
		return err
	}

	client := GetApiGWClient()

	for _, conn := range connections {
		_, err := client.PostToConnection(&apigatewaymanagementapi.PostToConnectionInput{ConnectionId: aws.String(conn.ConnectionID), Data: responseBytes})

		if err != nil {
			fmt.Printf("Error sending to %s: %v\n", conn.ConnectionID, err)
		}
	}

	return nil
}
