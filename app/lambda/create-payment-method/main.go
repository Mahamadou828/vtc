package main

import (
	"log"
	"vtc/app/lambda/create-payment-method/handler"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/business/v1/web"
	"vtc/foundation/config"
)

var app, err = config.NewApp()

func main() {
	if err != nil {
		log.Fatalf("failed to create a new app: %v", err)
	}

	awslambda.Start(web.NewHandler(handler.Handler, app))
}