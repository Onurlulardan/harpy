package resolver

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func TestLiteralResolutionErrorIsStableAndRedacted(t *testing.T) {
	t.Parallel()
	profile, _ := api.NewIntegerProfile(0, 10)
	fieldType, _ := api.NewIntegerFieldType(profile)
	const secret = "never-log-this-literal"
	_, err := ResolveLiterals(
		fieldType, api.OperatorEq, []api.Literal{mustTextLiteral(t, secret)},
	)
	requireResolutionCode(t, err, restqueryv1.IssueValueNotCanonical)
	formatted := fmt.Sprintf("%v|%+v|%#v|%q|%x", err, err, err, err, err)
	if strings.Contains(formatted, secret) || formatted != strings.Repeat(resolutionRedaction+"|", 4)+resolutionRedaction {
		t.Fatalf("resolution error formatting = %q", formatted)
	}
	if _, marshalErr := json.Marshal(err); marshalErr == nil ||
		strings.Contains(marshalErr.Error(), secret) {
		t.Fatalf("marshal error = %v", marshalErr)
	}
}

func TestLiteralResolutionErrorRejectsZeroAndTampering(t *testing.T) {
	t.Parallel()
	invalid := []LiteralResolutionError{
		{},
		{code: restqueryv1.IssueTypeMismatch},
		{code: restqueryv1.IssueCode("NOT_PUBLISHED"), valid: true},
	}
	for _, resolutionError := range invalid {
		if _, err := resolutionError.Code(); err == nil {
			t.Fatal("invalid resolution error exposed a code")
		}
	}
}
