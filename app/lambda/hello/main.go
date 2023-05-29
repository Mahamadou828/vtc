package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/business/v1/web"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

var appCfg, err = config.NewApp()

func main() {
	if err != nil {
		log.Fatalf("failed to create a new app: %v", err)
	}

	awslambda.Start(web.NewHandler(handler, appCfg))
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	return lambda.SendResponse(ctx, http.StatusOK, struct {
		Data []string `json:"data"`
	}{
		Data: os.Environ(),
	})
}
