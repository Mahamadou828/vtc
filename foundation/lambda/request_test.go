package lambda_test

import (
	"testing"
	"vtc/foundation/lambda"
)

const (
	//default json string for testing decode body
	jsonString = "{\"name\":\"test\"}"
)

func Test_DecodeBody(t *testing.T) {
	type Test struct {
		Name string `json:"name"`
	}

	var val Test

	t.Log("Given the need to decode json string")
	{
		if err := lambda.DecodeBody(jsonString, &val); err != nil {
			t.Logf("\t%s\t Test: \tShould be able to decode json body: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to decode json body:", success)
	}
}
