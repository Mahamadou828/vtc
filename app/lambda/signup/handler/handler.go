package handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	core "vtc/business/v1/core/user"
	"vtc/business/v1/data/models"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

func Handler(ctx context.Context, request events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	// Unmarshal body request and validate it
	var nu models.NewUserDTO
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
