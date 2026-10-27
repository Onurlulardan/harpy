package cursor

import (
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func FuzzCursorDecodeNeverPanicsOrLeaksNonQueryErrors(f *testing.F) {
	codec, context, token := encodedFixture(f)
	f.Add(token)
	f.Add("")
	f.Add("v1.current.invalid.invalid")
	f.Add(token + ".trailing")
	f.Fuzz(func(t *testing.T, candidate string) {
		claims, err := codec.Decode(candidate, context)
		if err == nil {
			if claims.Validate() != nil {
				t.Fatal("successful decode returned invalid claims")
			}
			return
		}
		if _, ok := err.(api.QueryError); !ok {
			t.Fatalf("decode returned non-query error %T: %v", err, err)
		}
	})
}
