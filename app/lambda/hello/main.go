package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"vtc/app/lambda/lambdasetup"
	"vtc/business/v1/sys/database"
	"vtc/business/v1/web"
	"vtc/foundation/lambda"
)

var client, err = database.NewClient(lambdasetup.SES, os.Getenv("DATABASE_POOL_NAME"), lambdasetup.SSLEnable)

func main() {
	if err != nil {
		log.Fatalf("Failed to init database connection: %v", err)
	}
	print(client.Timeout())
	//awslambda.Start(web.NewHandler(handler))
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *web.AppConfig) (events.APIGatewayProxyResponse, error) {
	return lambda.SendResponse(ctx, http.StatusOK, struct {
		Data []string `json:"data"`
	}{
		Data: os.Environ(),
	})
}
