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

func Handler(ctx context.Context, req events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	// Unmarshal body request and validate it
	var cred models.LoginDTO
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
