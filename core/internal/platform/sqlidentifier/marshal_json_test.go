package sqlidentifier_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
)

func TestTokenJSONEncodingFailsClosed(t *testing.T) {
	t.Parallel()
	token, err := sqlidentifier.MintStatic("forms_secret")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(struct {
		Identifier sqlidentifier.Token `json:"identifier"`
	}{Identifier: token})
	if err == nil {
		t.Fatalf("token crossed JSON boundary: %s", encoded)
	}
	if strings.Contains(err.Error(), "forms_secret") {
		t.Fatalf("JSON error exposed identifier: %v", err)
	}
}
