package web

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

const (
	EventFilePath        = "./event.local.json"
	AggregatorHeaderName = "aggregator"
)

type LambdaHandler func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Handler func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *config.App, trace *lambda.RequestTrace) (events.APIGatewayProxyResponse, error)

// NewHandler create a new LambdaHandler and pass it the default parameter
// NewHandler will also handle local testing by swapping the default request with event.local.json file content
func NewHandler(h Handler, cfg *config.App) LambdaHandler {
	//return the lambda handler
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		//handling local request if the app is run in local mode
		var err error
		if os.Getenv("APP_ENV") == "local" {
			//replace the passed request by a mock one before running the code
			request, err = GetLocalRequestEvent()
			if err != nil {
				//@todo handle the error with telemetry
				log.Fatalf("failed to extract local event: %v", err)
			}
		}

		//extract aggregator from header
		agg, ok := request.Headers[AggregatorHeaderName]
		if !ok {
			//@todo handle the error with telemetry
			log.Fatalf("aggregator code missing in hearder")
		}

		//Create a new request trace
		trace := lambda.RequestTrace{
			Now:        time.Now(),
			ID:         uuid.NewString(),
			Aggregator: agg,
		}

		//Put the new trace inside the context
		ctx := context.WithValue(context.Background(), lambda.CtxKey, &trace)

		return h(ctx, request, cfg, &trace)
	}
}

// GetLocalRequestEvent extract and parse local json file to mock event request
func GetLocalRequestEvent() (events.APIGatewayProxyRequest, error) {
	var event events.APIGatewayProxyRequest
	file, err := os.Open(EventFilePath)
	if err != nil {
		return event, fmt.Errorf("failed to open file: %v", err)
	}
	defer file.Close()

	decode := json.NewDecoder(file)

	if err := decode.Decode(&event); err != nil {
		return event, fmt.Errorf("failed to parse event file: %v", err)
	}

	return event, nil
}
