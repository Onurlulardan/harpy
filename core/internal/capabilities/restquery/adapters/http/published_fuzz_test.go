package httpquery

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func FuzzPublishedGETDecoderNeverPanics(f *testing.F) {
	for _, seed := range []string{
		"", "param[count]=1", "param[count]=%FF", "param[flags][01]=true",
		"param[count]=1&param%5Bcount%5D=2", "page[size]=500",
	} {
		f.Add(seed)
	}
	decoder := NewPublishedGETDecoder()
	f.Fuzz(func(t *testing.T, raw string) {
		bound, err := decoder.Decode(raw, publishedTestManifest(t))
		requirePublishedFuzzOutcome(t, bound.Validate(), err)
	})
}

func FuzzPublishedPOSTDecoderNeverPanics(f *testing.F) {
	for _, seed := range [][]byte{
		[]byte(`{}`),
		[]byte(`{"params":{"count":"1"}}`),
		[]byte(`{"params":{"count":"1","flags":[true,false]}}`),
		[]byte(`{"params":{"count":"1","count":"2"}}`),
		{0xff, 0x00, '{'},
	} {
		f.Add(seed)
	}
	decoder := NewPublishedPOSTDecoder()
	f.Fuzz(func(t *testing.T, document []byte) {
		bound, err := decoder.Decode(document, publishedTestManifest(t))
		requirePublishedFuzzOutcome(t, bound.Validate(), err)
	})
}

func requirePublishedFuzzOutcome(t *testing.T, validationErr, decodeErr error) {
	t.Helper()
	if decodeErr == nil {
		if validationErr != nil {
			t.Fatalf("decoder returned invalid bound parameters: %v", validationErr)
		}
		return
	}
	var queryError api.QueryError
	if !errors.As(decodeErr, &queryError) {
		t.Fatalf("decoder error type = %T", decodeErr)
	}
}
