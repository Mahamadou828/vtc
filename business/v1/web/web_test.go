package web_test

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"testing"
	"time"
	"vtc/business/v1/web"
	"vtc/foundation/lambda"
)

const (
	success = "\u2713"
	failure = "\u2717"
)

func TestWeb(t *testing.T) {
	t.Log("Given the need to create new handler that match the aws lambda signature")
	{
		handler := func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *web.AppConfig) (events.APIGatewayProxyResponse, error) {
			t.Log("Given the need to pass a context with request trace")
			{
				r, err := lambda.GetRequestTrace(ctx)
				if err != nil {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, err)
				}
				if len(r.ID) == 0 {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, fmt.Errorf("uuid inside request trace should not be empty"))
				}
				if r.Now != time.Now() {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, fmt.Errorf("now date is invalid"))
				}
				t.Logf("\t%s\t Test: \tShould receive a request trace", success)
			}

			return events.APIGatewayProxyResponse{}, nil
		}

		web.NewHandler(handler, nil)(events.APIGatewayProxyRequest{})
	}
}
