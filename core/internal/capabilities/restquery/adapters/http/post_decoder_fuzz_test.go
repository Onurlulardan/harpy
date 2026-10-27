package httpquery

import (
	"errors"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
)

func FuzzPOSTDecoderNeverPanics(f *testing.F) {
	seeds := [][]byte{
		[]byte(`{}`),
		[]byte(`{"where":{"field":"id","op":"eq","value":"1"}}`),
		[]byte(`{"aggregate":{"measures":[{"name":"rows","function":"count"}]}}`),
		[]byte(`{"x":1,"x":2}`),
		{0xff, 0x00, '{'},
	}
	for _, seed := range seeds {
		f.Add(seed)
	}
	decoder := NewPOSTDecoder()
	f.Fuzz(func(t *testing.T, document []byte) {
		request, err := decoder.Decode(document)
		if err != nil {
			var queryError api.QueryError
			if !errors.As(err, &queryError) {
				t.Fatalf("Decode() error type = %T", err)
			}
			return
		}
		if err := request.Validate(); err != nil {
			t.Fatalf("Decode() returned invalid request: %v", err)
		}
	})
}
