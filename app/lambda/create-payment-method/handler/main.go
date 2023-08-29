package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	core "vtc/business/v1/core/user"
	model "vtc/business/v1/data/models/user"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	var data model.NewPaymentMethodDTO

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
