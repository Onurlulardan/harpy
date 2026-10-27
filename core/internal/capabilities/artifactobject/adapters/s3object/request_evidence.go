package s3object

import (
	"errors"
	"net/http"

	awsmiddleware "github.com/aws/aws-sdk-go-v2/aws/middleware"
	"github.com/aws/smithy-go/middleware"
	smithyhttp "github.com/aws/smithy-go/transport/http"
)

func metadataRequestID(value middleware.Metadata) (string, error) {
	requestID, ok := awsmiddleware.GetRequestIDMetadata(value)
	if !ok || requestID == "" {
		return "", ErrProviderEvidence
	}
	return requestID, nil
}

func errorRequestID(err error) string {
	var response *smithyhttp.ResponseError
	if !errors.As(err, &response) || response.HTTPResponse() == nil {
		return ""
	}
	return response.HTTPResponse().Header.Get("x-amz-request-id")
}

func responseStatus(err error) int {
	var response *smithyhttp.ResponseError
	if errors.As(err, &response) {
		return response.HTTPStatusCode()
	}
	return 0
}

func provesAbsence(err error) bool {
	status := responseStatus(err)
	return status == http.StatusNotFound || status == http.StatusGone
}

func knownNoWrite(err error) bool {
	switch responseStatus(err) {
	case http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden,
		http.StatusNotFound, http.StatusMethodNotAllowed, http.StatusLengthRequired,
		http.StatusRequestEntityTooLarge, http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType, http.StatusUnprocessableEntity:
		return true
	default:
		return false
	}
}
