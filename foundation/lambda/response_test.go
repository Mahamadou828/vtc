package lambda_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/http"
	"testing"
	"time"

	"vtc/foundation/lambda"
)

func Test_SendResponse(t *testing.T) {
	t.Log("Given the need to send response conform to the Api Proxy Response format")
	{
		trace := lambda.RequestTrace{
			ID:         uuid.NewString(),
			Now:        time.Now(),
			StatusCode: 200,
			Aggregator: "test",
		}

		ctx := context.WithValue(context.Background(), lambda.CtxKey, &trace)

		data := struct {
			Name string `json:"name"`
		}{
			Name: "test",
		}

		_, err := lambda.SendResponse(ctx, http.StatusOK, data)
		if err != nil {
			//@todo test if resp if in the good format
			t.Logf("\t%s\t Test: \tShould be able to create response: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to create response:", success)
	}
}

func Test_SendError(t *testing.T) {
	t.Log("Given the need to send error response conform to the Api Proxy Response format")
	{
		//@todo test if resp if in the good format
		_, err := lambda.SendError(context.Background(), http.StatusBadRequest, fmt.Errorf("failed response"))
		if err != nil {
			t.Logf("\t%s\t Test: \tShould be able to generate error response: %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to generate error response :", success)
	}
}
