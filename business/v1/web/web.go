package web

import (
	"context"
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

type Env struct {
	Cognito struct {
		ClientID string `conf:"env:COGNITO_CLIENT_ID,required"`
	}
	Stripe struct {
		Key string `conf:"env:STRIPE_KEY,required"`
	}
}

type AppConfig struct {
	DBClient   *mongo.Database
	AWSSession *session.Session
	Env        Env
}

type LambdaHandler func(request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error)

type Handler func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *AppConfig, trace *lambda.RequestTrace) (events.APIGatewayProxyResponse, error)

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
		//extract aggregator from header
		agg, ok := request.Headers["aggregator"]
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
