package s3blob

import (
	"errors"
	"net/http"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func requestID(metadata middleware.Metadata) string {
	value, ok := awsmiddleware.GetRequestIDMetadata(metadata)
	if !ok {
		return ""
	}
	return value
}

func errorRequestID(err error) string {
	var response *smithyhttp.ResponseError
	if !errors.As(err, &response) || response.HTTPResponse() == nil {
		return ""
	}
	return response.HTTPResponse().Header.Get("x-amz-request-id")
}

func provesAbsence(err error) bool {
	var response *smithyhttp.ResponseError
	if !errors.As(err, &response) {
		return false
	}
	return response.HTTPStatusCode() == http.StatusNotFound ||
		response.HTTPStatusCode() == http.StatusGone
}
