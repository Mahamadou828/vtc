package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"vtc/business/v1/core/auth"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/app/lambda/lambdasetup"
	model "vtc/business/v1/data/models/auth"
	"vtc/business/v1/sys/database"
	"vtc/business/v1/sys/validate"
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
	// Unmarshal body request and validate it
	var nu model.NewUserDTO
	if err := lambda.DecodeBody(request.Body, &nu); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to decode body: %v", err))
	}

	//validate body
	if err := validate.Check(&nu); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to validate body: %v", err))
	}

	//create the new user
	u, err := auth.SignUp(ctx, nu, cfg, t.Aggregator, t.Now)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to create new user: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusCreated, u)
}
