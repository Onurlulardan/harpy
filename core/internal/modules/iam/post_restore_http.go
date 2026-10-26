package iam

import (
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/postrestore"
)

type PostRestoreHTTP struct {
	service *postrestore.Service
	auth    *AuthHTTP
	report  HTTPErrorSink
}

func newPostRestoreHTTP(
	service *postrestore.Service, auth *AuthHTTP, report HTTPErrorSink,
) (*PostRestoreHTTP, error) {
	if service == nil || auth == nil || report == nil {
		return nil, fmt.Errorf("IAM post-restore HTTP dependencies are required")
	}
	return &PostRestoreHTTP{service: service, auth: auth, report: report}, nil
}
