package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"os"
	"vtc/business/v1/web"
	"vtc/foundation/lambda"
)

func main() {
	awslambda.Start(web.NewHandler(handler))
}

func handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *web.AppConfig) (events.APIGatewayProxyResponse, error) {
	return lambda.SendResponse(ctx, http.StatusOK, struct {
		Data []string `json:"data"`
	}{
		Data: os.Environ(),
	})
}
