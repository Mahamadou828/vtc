package main

import (
	"log"

	awslambda "github.com/aws/aws-lambda-go/lambda"
	"vtc/app/lambda/create-payment/handler"
	"vtc/business/v1/web"
	"vtc/foundation/config"
)

var app, err = config.NewApp()

func main() {
	if err != nil {
		log.Fatalf("failed to create a new app: [%v]", err)
	}

	awslambda.Start(web.NewHandler(handler.Handler, app))
}
