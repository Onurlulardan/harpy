package api

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	restqueryv1 "github.com/Onurlulardan/harpy/contracts/restquery/v1"
	runtimev1 "github.com/Onurlulardan/harpy/contracts/runtime/v1"
)

func TestQueryErrorFormattingOnlyExposesOuterCode(t *testing.T) {
	queryError, err := NewQueryError(
		restqueryv1.ErrorQueryInvalid,
		[]runtimev1.Issue{{Path: "/secret/path", Code: "FIELD_UNKNOWN"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	want := string(restqueryv1.ErrorQueryInvalid)
	for _, formatted := range []string{
		queryError.Error(), queryError.String(), queryError.GoString(),
		fmt.Sprint(queryError), fmt.Sprintf("%s", queryError),
		fmt.Sprintf("%q", queryError), fmt.Sprintf("%#v", queryError),
	} {
		if formatted != want {
			t.Fatalf("formatting = %q, want %q", formatted, want)
		}
	}
	_, err = json.Marshal(queryError)
	if err == nil {
		t.Fatal("opaque query error was silently marshaled")
	}
	if strings.Contains(err.Error(), "secret") {
		t.Fatalf("marshal error leaked issue path: %v", err)
	}
}
