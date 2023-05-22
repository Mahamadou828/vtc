package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/app/lambda/lambdasetup"
	"vtc/business/v1/sys/database"
	"vtc/business/v1/web"
	"vtc/foundation/lambda"
)

var client, err = database.NewClient(lambdasetup.DatabaseConfig)

func main() {
	if err != nil {
		log.Fatalf("Failed to init database connection: %v", err)
	}
	awslambda.Start(web.NewHandler(handler, client))
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *web.AppConfig, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	return lambda.SendResponse(ctx, http.StatusOK, struct {
		Data []string `json:"data"`
	}{
		Data: os.Environ(),
	})
}
