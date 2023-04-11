package web

import (
	"context"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"vtc/foundation/lambda"
)

type AppConfig struct {
}

type LambdaHandler func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Handler func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *AppConfig) (events.APIGatewayProxyResponse, error)

func NewHandler(h Handler) LambdaHandler {
	//Create a new AppConfig
	cfg := &AppConfig{}

	//Create a new request trace
	trace := lambda.RequestTrace{
		Now: time.Now(),
		ID:  uuid.NewString(),
	}

	//Put the new trace inside the context
	ctx := context.WithValue(context.Background(), lambda.CtxKey, trace)

	//return the lambda handler
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		return h(ctx, request, cfg)
	}
}
