package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
)

func main() {
	lambda.Start(HandleRequest)
}

func HandleRequest() (events.APIGatewayProxyResponse, error) {
	return events.APIGatewayProxyResponse{
		StatusCode: http.StatusOK,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Methods": "GET,POST,OPTIONS,PUT,PATCH,DELETE",
			"Access-Control-Allow-Origin":  "*",
		},
		Body: "Hello from the other side",
	}, nil
}
