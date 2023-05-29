package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	core "vtc/business/v1/core/auth"
	model "vtc/business/v1/data/models/auth"
	"vtc/business/v1/sys/validate"
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
	u, err := core.SignUp(ctx, nu, cfg, t.Aggregator, t.Now)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to create new user: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusCreated, u)
}
