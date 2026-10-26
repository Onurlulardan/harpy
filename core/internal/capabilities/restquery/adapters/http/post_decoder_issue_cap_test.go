package httpquery

import (
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
)

func TestPOSTDecoderSortsAndCapsSemanticIssues(t *testing.T) {
	t.Parallel()
	members := make([]string, 25)
	for index := range members {
		members[index] = fmt.Sprintf(`"unknown_%02d":true`, 24-index)
	}
	document := []byte(`{` + strings.Join(members, ",") + `}`)
	errorValue := requireQueryError(t, document, restqueryv1.ErrorQueryInvalid,
		expectedUnknownIssues(20))
	issues := errorValue.Issues()
	if issues[0].Path != "/unknown_00" || issues[19].Path != "/unknown_19" {
		t.Fatalf("capped issue range = %q..%q", issues[0].Path, issues[19].Path)
	}
}

func expectedUnknownIssues(count int) []expectedQueryIssue {
	result := make([]expectedQueryIssue, count)
	for index := range result {
		result[index] = expectedQueryIssue{
			path: fmt.Sprintf("/unknown_%02d", index),
			code: restqueryv1.IssueMemberUnknown,
		}
	}
	return result
}
