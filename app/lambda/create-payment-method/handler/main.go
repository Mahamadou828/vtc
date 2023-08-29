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
	var data models.NewPaymentMethodDTO

	if err := lambda.DecodeBody(req.Body, &data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to decode body: %v", err))
	}

	if err := validate.Check(&data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("invalid request body: %v", err))
	}

	pm, err := core.CreatePaymentMethod(ctx, data, cfg, t.Now)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to create a new payment method: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusCreated, pm)
}
