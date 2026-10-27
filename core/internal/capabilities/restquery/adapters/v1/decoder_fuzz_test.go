package v1query

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func FuzzV1DecoderFailsClosed(f *testing.F) {
	for _, seed := range []string{
		"", "page=2&limit=10", "filters[status]=OPEN",
		"filters[tags][]=red&filters[tags][]=blue", "search=%zz", "unknown=x",
	} {
		f.Add(seed)
	}
	decoder, err := newDecoder(staticManifestSource{snapshot: testManifest()})
	if err != nil {
		f.Fatalf("NewDecoder() error = %v", err)
	}
	f.Fuzz(func(t *testing.T, raw string) {
		result, decodeErr := decoder.Decode(raw)
		if decodeErr == nil {
			if err := result.Validate(); err != nil {
				t.Fatalf("successful Decode returned invalid result: %v", err)
			}
			return
		}
		var queryError api.QueryError
		if !errors.As(decodeErr, &queryError) {
			t.Fatalf("client input returned internal error %T: %v", decodeErr, decodeErr)
		}
	})
}
