package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
	"vtc/business/v1/core/provider"
	"vtc/business/v1/data/models"
	"vtc/business/v1/sys/validate"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

func Handler(ctx context.Context, req events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	var data models.CreatePaymentDTO

	if err := lambda.DecodeBody(req.Body, &data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to decode request body: %v", err))
	}

	if err := validate.Check(&data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("invalid body: %v", err))
	}

	charge, err := provider.CreatePayment(ctx, data, cfg, t.Now)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to create payment: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusCreated, charge)
}
