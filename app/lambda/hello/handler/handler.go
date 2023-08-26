package handler

import (
	"context"
	"net/http"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	return lambda.SendResponse(ctx, http.StatusOK, struct {
		Data []string `json:"data"`
	}{
		Data: os.Environ(),
	})
}
