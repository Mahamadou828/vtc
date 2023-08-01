package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"vtc/business/v1/core/provider"
	mOffer "vtc/business/v1/data/models/offer"
	"vtc/business/v1/sys/validate"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/business/v1/web"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

var app, err = config.NewApp()

func main() {
	if err != nil {
		log.Fatalf("failed to create a new app: %v", err)
	}
	awslambda.Start(web.NewHandler(handler, app))
}

func handler(ctx context.Context, req events.APIGatewayProxyRequest, cfg *config.App, t *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
	var data mOffer.GetOfferDTO

	if err := lambda.DecodeBody(req.Body, &data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to decode body: %v", err))
	}
	if err := validate.Check(&data); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("invalid body: %v", err))
	}
	if err := validate.CheckID(data.UserID); err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("invalid user id"))
	}

	offers, err := provider.GetOffers(ctx, data, cfg, t.Aggregator, t.Now)
	if err != nil {
		return lambda.SendError(ctx, http.StatusBadRequest, fmt.Errorf("failed to get offer: %v", err))
	}

	return lambda.SendResponse(ctx, http.StatusOK, offers)
}
