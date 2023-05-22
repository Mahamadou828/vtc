package lambda

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"

	"github.com/aws/aws-lambda-go/events"
)

var (
	ContentTypeHeaderMissingError         = errors.New("content type header missing")
	ContentTypeHeaderNotMultipartError    = errors.New("content type header not multipart error")
	ContentTypeHeaderMissingBoundaryError = errors.New("content type header missing boundary error")
)

// DecodeBody decode the body from base64 to json and parse it into the given dest
func DecodeBody(body string, val any) error {
	dBytes, err := base64.StdEncoding.DecodeString(body)
	if err != nil {
		return fmt.Errorf("failed to decode body: %v", err)
	}

	decoder := json.NewDecoder(bytes.NewReader(dBytes))
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(val); err != nil {
		return err
	}

	return nil
}

func standardHeader(header map[string]string) http.Header {
	h := http.Header{}
	for k, v := range header {
		h.Add(strings.TrimSpace(k), v)
	}
	return h
}

// NewReaderMultipart create a new multipart reader from a aws api gateway request
func NewReaderMultipart(req events.APIGatewayProxyRequest) (*multipart.Reader, error) {
	headers := standardHeader(req.Headers)
	ct := headers.Get("content-type")
	if len(ct) == 0 {
		return nil, ContentTypeHeaderMissingError
	}

	mediatype, params, err := mime.ParseMediaType(ct)
	if err != nil {
		return nil, fmt.Errorf("unable to parse mediatype: %v", err)
	}

	if strings.Index(strings.ToLower(strings.TrimSpace(mediatype)), "multipart/") != 0 {
		return nil, ContentTypeHeaderNotMultipartError
	}

	paramsInsensitiveKeys := standardHeader(params)
	boundary := paramsInsensitiveKeys.Get("boundary")
	if len(boundary) == 0 {
		return nil, ContentTypeHeaderMissingBoundaryError
	}

	if req.IsBase64Encoded {
		decoded, err := base64.StdEncoding.DecodeString(req.Body)
		if err != nil {
			return nil, fmt.Errorf("can't decode body: %v", err)
		}
		return multipart.NewReader(bytes.NewReader(decoded), boundary), nil
	}

	return multipart.NewReader(strings.NewReader(req.Body), boundary), nil
}
