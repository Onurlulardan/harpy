package passwordblocklistv1

import (
	"crypto/sha256"
	"errors"
	"testing"
)

func TestParseValidatesAndSearchesCanonicalArtifact(t *testing.T) {
	first := sha256.Sum256([]byte("first password"))
	second := sha256.Sum256([]byte("second password"))
	content, err := Encode(KindCommon, [][32]byte{second, first})
	if err != nil {
		t.Fatal(err)
	}
	artifact, err := Parse(content, KindCommon)
	if err != nil {
		t.Fatal(err)
	}
	count, err := artifact.EntryCount()
	if err != nil || count != 2 {
		t.Fatalf("entry count=%d error=%v", count, err)
	}
	found, err := artifact.Contains(second)
	if err != nil || !found {
		t.Fatalf("known digest found=%t error=%v", found, err)
	}
	unknown := sha256.Sum256([]byte("unknown password"))
	found, err = artifact.Contains(unknown)
	if err != nil || found {
		t.Fatalf("unknown digest found=%t error=%v", found, err)
	}
}

func TestParseRejectsKindShapeAndOrderingDrift(t *testing.T) {
	one := sha256.Sum256([]byte("one"))
	two := sha256.Sum256([]byte("two"))
	valid, err := Encode(KindCommon, [][32]byte{one, two})
	if err != nil {
		t.Fatal(err)
	}
	tests := map[string][]byte{
		"truncated": append([]byte(nil), valid[:len(valid)-1]...),
		"reserved":  mutateArtifact(valid, 11, 1),
		"profile":   mutateArtifact(valid, 9, 2),
		"duplicate": duplicateFirstRecord(valid),
	}
	for name, content := range tests {
		t.Run(name, func(t *testing.T) {
			if _, err := Parse(content, KindCommon); !errors.Is(err, ErrInvalidArtifact) {
				t.Fatalf("invalid artifact error=%v", err)
			}
		})
	}
	if _, err := Parse(valid, KindCompromised); !errors.Is(err, ErrInvalidArtifact) {
		t.Fatalf("cross-kind artifact error=%v", err)
	}
}

func mutateArtifact(source []byte, index int, value byte) []byte {
	result := append([]byte(nil), source...)
	result[index] = value
	return result
}

func duplicateFirstRecord(source []byte) []byte {
	result := append([]byte(nil), source...)
	copy(result[headerSize+digestSize:], result[headerSize:headerSize+digestSize])
	return result
}
