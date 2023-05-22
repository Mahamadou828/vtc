package main

import (
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"testing"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestCognitoPreSignupFunc(t *testing.T) {
	t.Log("Given the need to automatically confirm a account on aws cognito")
	{
		var event events.CognitoEventUserPoolsPreSignup
		res, err := handler(event)
		if err != nil {
			t.Fatalf("\t%s\t Test: \tShould be able to auto confirm a account: %v", failure, err)
		}

		if !res.Response.AutoConfirmUser {
			t.Fatalf("\t%s\t Test: \tShould be able to auto confirm a account: %v", failure, fmt.Errorf("return false for autom confirm user"))
		}

		t.Logf("\t%s\t Test: \tShould be able to auto confirm a account", success)
	}
}
