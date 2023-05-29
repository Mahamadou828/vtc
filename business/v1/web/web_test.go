package web_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"vtc/business/v1/web"
	"vtc/foundation/config"
	"vtc/foundation/lambda"
)

const (
	success        = "\u2713"
	failure        = "\u2717"
	aggregatorName = "test"
)

func Test_NewHandler(t *testing.T) {
	t.Log("Given the need to create new handler that match the aws lambda signature")
	{
		handler := func(ctx context.Context, request events.APIGatewayProxyRequest, cfg *config.App, trace *lambda.RequestTrace) (events.APIGatewayProxyResponse, error) {
			t.Log("Given the need to pass a context with request trace")
			{
				r, err := lambda.GetRequestTrace(ctx)
				if err != nil {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, err)
				}
				if r.ID != trace.ID {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, fmt.Errorf("uuid inside request trace is not equal to the passed trace"))
				}
				if r.Now != trace.Now {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, fmt.Errorf("now date is invalid"))
				}
				if r.Aggregator != aggregatorName {
					t.Fatalf("\t%s\t Test: \tShould receive a request trace: %v", failure, fmt.Errorf("aggregator is invalid"))
				}
				t.Logf("\t%s\t Test: \tShould receive a request trace", success)
			}

			return events.APIGatewayProxyResponse{}, nil
		}

		web.NewHandler(handler, nil)(events.APIGatewayProxyRequest{Headers: map[string]string{"aggregator": aggregatorName}})
	}
}
