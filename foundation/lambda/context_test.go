package lambda_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"testing"
	"time"
	"vtc/foundation/lambda"
)

const (
	success = "\u2713"
	failure = "\u2717"

	aggregatorName = "test"
)

var (
	ctx   context.Context
	trace lambda.RequestTrace
)

func TestMain(m *testing.M) {
	trace = lambda.RequestTrace{
		ID:         uuid.NewString(),
		Now:        time.Now(),
		Aggregator: aggregatorName,
	}

	ctx = context.WithValue(context.Background(), lambda.CtxKey, &trace)
}

func Test_GetRequestTrace(t *testing.T) {
	t.Log("Given the need to retrieve a request trace")
	{
		ctxTrace, err := lambda.GetRequestTrace(ctx)
		if err != nil {
			t.Logf("\t%s\t Test: \tShould be able to get the request trace: %v", failure, err)
		}
		if ctxTrace.ID != trace.ID {
			t.Logf("\t%s\t Test: \tRequest trace id are not identiqual : %v", failure, err)
		}
		t.Logf("\t%s\t Test: \tShould be able to get the request trace:", success)
	}
}

func Test_GetTraceID(t *testing.T) {
	t.Log("Given the need to retrieve a request trace ID")
	{
		if id := lambda.GetTraceID(ctx); id != trace.ID {
			t.Logf("\t%s\t Test: \tShould be able to retrieve request trace : %v", failure, fmt.Errorf("request trace id not found"))
		}
		t.Logf("\t%s\t Test: \tShould be able to to retrieve request trace:", success)
	}
}
