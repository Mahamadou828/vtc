package web

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"os"
	"time"
	"vtc/app/tools/config"

	"github.com/aws/aws-lambda-go/events"
	"github.com/google/uuid"
	"vtc/foundation/lambda"
)

const (
	EventFilePath        = "./event.local.json"
	AggregatorHeaderName = "aggregator"
)

// Env defines all environment variable needed to run the application
type Env struct {
	Cognito struct {
		ClientID string `conf:"env:COGNITO_CLIENT_ID,required"`
	}
	Stripe struct {
		Key string `conf:"env:STRIPE_KEY,required"`
	}
}

// AppConfig defines all the necessary dependencies to run the application
type AppConfig struct {
	DBClient   *mongo.Database
	AWSSession *session.Session
	Env        Env
}

type LambdaHandler func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Handler func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *AppConfig, trace *lambda.RequestTrace) (events.APIGatewayProxyResponse, error)

// NewHandler create a new LambdaHandler and pass it the default parameter
// NewHandler will also handle local testing by swapping the default request with event.local.json file content
func NewHandler(h Handler, client *mongo.Database) LambdaHandler {
	//init a new aws session
	sess, err := session.NewSession(
		&aws.Config{
			Region:                        aws.String(os.Getenv("AWS_REGION")),
			CredentialsChainVerboseErrors: aws.Bool(true),
		},
	)
	if err != nil {
		//@todo handle the error with telemetry
		log.Fatalf("failed to init new aws session: %v", err)
	}

	//extract all env variable
	var env Env
	if err := config.ParseEnv(&env); err != nil {
		//@todo handle the error with telemetry
		log.Fatalf("failed to extract required env config: %v", err)
	}

	//Create a new AppConfig
	cfg := &AppConfig{
		DBClient:   client,
		AWSSession: sess,
		Env:        env,
	}

	//return the lambda handler
	return func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
		//handling local request if the app is run in local mode
		if os.Getenv("APP_ENV") == "local" {
			//replace the passed request by a mock one before running the code
			request, err = getLocalRequestEvent()
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
		ctx := context.WithValue(context.Background(), lambda.CtxKey, trace)

		return h(ctx, request, cfg, nil)
	}
}

// getLocalRequestEvent extract and parse local json file to mock event request
func getLocalRequestEvent() (events.APIGatewayProxyRequest, error) {
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
