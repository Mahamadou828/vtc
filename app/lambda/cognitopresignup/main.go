package main

import (
	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
)

func main() {
	awslambda.Start(handler)
}

// handler is a function that automatically confirm a user account inside aws cognito
// for now we will not implement verification of phone number or email and just confirm the account
func handler(event events.CognitoEventUserPoolsPreSignup) (events.CognitoEventUserPoolsPreSignup, error) {
	event.Response.AutoConfirmUser = true
	return event, nil
}
