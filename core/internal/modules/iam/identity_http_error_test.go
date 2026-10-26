package iam

import (
	"errors"
	"testing"

	userapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identity/user"
	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
)

func TestIdentityApplicationErrorIsClosedAndPreservesPublishedQueryFailure(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want error
	}{
		{"nil", nil, nil},
		{"owned invalid", ErrIdentityHTTPInvalid, ErrIdentityHTTPInvalid},
		{"owned forbidden", ErrIdentityHTTPForbidden, ErrIdentityHTTPForbidden},
		{"user command", userapp.ErrInvalidCommand, ErrIdentityHTTPInvalid},
		{"user query", userapp.ErrInvalidQuery, ErrIdentityHTTPInvalid},
		{"unknown", errors.New("storage detail"), ErrIdentityHTTPUnavailable},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := identityApplicationError(test.err)
			if test.want == nil && actual != nil || test.want != nil && !errors.Is(actual, test.want) {
				t.Fatalf("error=%v want=%v", actual, test.want)
			}
		})
	}
	_, failure := queryapp.DecodePOST([]byte(`{"unknown":true}`))
	if failure == nil {
		t.Fatal("invalid query fixture did not produce a published failure")
	}
	actual := identityApplicationError(failure)
	var published queryapp.Failure
	if !errors.As(actual, &published) || published.Code() != "QUERY_INVALID" {
		t.Fatalf("query error was normalized away: %v", actual)
	}
}
