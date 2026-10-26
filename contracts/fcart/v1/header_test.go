package fcartv1

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

func TestHeaderRoundTripForEveryFormat(t *testing.T) {
	for _, format := range []Format{FormatCSV, FormatJSONL, FormatXLSX, FormatPDF} {
		t.Run(string(format), func(t *testing.T) {
			header, err := NewHeader(uuid.New(), 1, format)
			if err != nil {
				t.Fatal(err)
			}
			encoded, err := header.encode()
			if err != nil {
				t.Fatal(err)
			}
			decoded, err := decodeHeader(encoded)
			if err != nil || decoded != header {
				t.Fatalf("decoded = %#v, %v", decoded, err)
			}
		})
	}
}

func TestHeaderRejectsNonCanonicalValues(t *testing.T) {
	valid := fixtureHeader(t)
	for name, mutate := range map[string]func(*Header){
		"version":  func(value *Header) { value.Version = "FCART2" },
		"artifact": func(value *Header) { value.ArtifactID = uuid.Nil },
		"attempt":  func(value *Header) { value.AttemptGeneration = 0 },
		"format":   func(value *Header) { value.Format = "zip" },
		"chunk":    func(value *Header) { value.ChunkSize-- },
	} {
		t.Run(name, func(t *testing.T) {
			changed := valid
			mutate(&changed)
			if !errors.Is(changed.Validate(), ErrInvalidHeader) {
				t.Fatal("non-canonical header accepted")
			}
		})
	}
}

func TestDEKRequiresExactlyNonzero256Bits(t *testing.T) {
	for _, raw := range [][]byte{nil, make([]byte, DEKBytes-1), make([]byte, DEKBytes)} {
		if _, err := NewDEK(raw); !errors.Is(err, ErrInvalidDEK) {
			t.Fatalf("invalid DEK length %d accepted", len(raw))
		}
	}
	if fixtureDEK(t).Validate() != nil {
		t.Fatal("valid DEK rejected")
	}
}
