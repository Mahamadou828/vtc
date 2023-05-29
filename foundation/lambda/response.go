package lambda

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"net/http"
)

// SendResponse format the response to match the api proxy response format
func SendResponse(ctx context.Context, status int, data any) (events.APIGatewayProxyResponse, error) {
	trace, err := GetRequestTrace(ctx)
	if err != nil {
		return SendError(ctx, http.StatusInternalServerError, fmt.Errorf("failed to retrieve request trace: %v", err))
	}

	b, err := json.Marshal(data)
	if err != nil {
		return SendError(ctx, http.StatusInternalServerError, errors.New("can't marshal response"))
	}

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Methods": "GET,POST,OPTIONS,PUT,PATCH,DELETE",
			"Access-Control-Allow-Origin":  "*",
			"Content-Type":                 "application/json",
			"TraceID":                      trace.ID,
			"aggregator":                   trace.Aggregator,
		},
		Body: string(b),
	}, nil
}

// SendError format an error response to match the api proxy response spec
func SendError(ctx context.Context, status int, err error) (events.APIGatewayProxyResponse, error) {
	data := struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	}

	b, _ := json.Marshal(data)

	return events.APIGatewayProxyResponse{
		StatusCode: status,
		Headers: map[string]string{
			"Access-Control-Allow-Headers": "*",
			"Access-Control-Allow-Methods": "GET,POST,OPTIONS,PUT,PATCH,DELETE",
			"Access-Control-Allow-Origin":  "*",
			"Content-Type":                 "application/json",
			"TraceID":                      GetTraceID(ctx),
		},
		Body: string(b),
	}, nil
}
