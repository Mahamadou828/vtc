package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	core "vtc/business/v1/core/auth"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	model "vtc/business/v1/data/models/auth"
	"vtc/business/v1/web"
	"vtc/foundation/lambda"
)

var appCfg, err = config.NewApp()

func main() {
	if err != nil {
		log.Fatalf("failed to create a new app: %v", err)
	}

	awslambda.Start(web.NewHandler(handler, appCfg))
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	// Unmarshal body request and validate it
	var cred model.LoginDTO
	if err := lambda.DecodeBody(req.Body, &cred); err != nil {
		return lambda.SendError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to decode body: %v", err))
	}

	//validate body
	if err := validate.Check(&cred); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("invalid body: %v", err))
	}

	//log the given user and return the session
	sess, err := core.Login(ctx, cred, cfg, t.Aggregator)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to log user: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusOK, sess)
}
