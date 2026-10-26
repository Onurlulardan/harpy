package billing

import (
	"errors"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	billingapi "github.com/Onurlulardan/harpy/core/internal/modules/billing/application/httpapi"
)

func TestPlatformReportKeepsCanonicalQueryError(t *testing.T) {
	_, queryErr := billingapi.NewQueryDecoder().DecodeGET("where=x")
	if queryErr == nil {
		t.Fatal("invalid REST Query was accepted")
	}
	normalized := normalizePlatformReportError(queryErr)
	_, code, _, valid := billingapi.QueryErrorDetails(normalized)
	if !valid || code != string(restqueryv1.ErrorQueryInvalid) {
		t.Fatalf("normalized query error = %v", normalized)
	}
}

func TestPlatformReportHidesUnexpectedFailure(t *testing.T) {
	if err := normalizePlatformReportError(errors.New("storage detail")); !errors.Is(err, billingapi.ErrUnavailable) {
		t.Fatalf("normalized error = %v", err)
	}
}
