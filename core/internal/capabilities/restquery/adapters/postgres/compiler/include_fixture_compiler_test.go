package compiler

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/cursor"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/configfixture"
)

func includeCompiler(t *testing.T) Compiler {
	t.Helper()
	key := func(value byte) string {
		return base64.RawURLEncoding.EncodeToString(bytes.Repeat([]byte{value}, 32))
	}
	document := fmt.Sprintf(
		`{"version":1,"current":{"kid":"current-1","key":"%s"},`+
			`"previous":{"kid":"previous-1","key":"%s"}}`, key(3), key(4),
	)
	t.Setenv("HARPY_ENVIRONMENT", "development")
	t.Setenv("HARPY_POSTGRES_ENDPOINT_KIND", "direct_postgres")
	t.Setenv("HARPY_POSTGRES_DSN",
		"postgresql://user@example.invalid/harpy?sslmode=disable")
	t.Setenv("HARPY_REST_QUERY_CURSOR_KEYRING", document)
	configfixture.SetAPI(t)
	loaded, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	keyring, err := loaded.RestQueryCursorKeyring()
	if err != nil {
		t.Fatal(err)
	}
	codec, err := cursor.NewCodec(keyring)
	if err != nil {
		t.Fatal(err)
	}
	compiler, err := New(codec)
	if err != nil {
		t.Fatal(err)
	}
	return compiler
}
