package archiveartifact

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/sha256"
	"io"
	"testing"

	archiveapp "github.com/Onurlulardan/harpy/core/internal/modules/documents/application/archive"
	"github.com/Onurlulardan/harpy/core/internal/modules/documents/domain"
)

type zipFixtureSource map[domain.ID][]byte

func (source zipFixtureSource) Open(_ context.Context,
	entry archiveapp.Entry,
) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(source[entry.VersionID])), nil
}

func TestStreamingZIPHasExactPrecomputedSizeAndSafeNames(t *testing.T) {
	firstID := zipFixtureID(t, "018f0000-0000-7000-8000-000000000301")
	firstVersion := zipFixtureID(t, "018f0000-0000-7000-8000-000000000302")
	secondID := zipFixtureID(t, "018f0000-0000-7000-8000-000000000303")
	secondVersion := zipFixtureID(t, "018f0000-0000-7000-8000-000000000304")
	first, second := []byte("first evidence"), []byte("second evidence")
	entries := []archiveapp.Entry{
		{ArchiveEntry: domain.ArchiveEntry{DocumentID: firstID, VersionID: firstVersion,
			Order: 0, PlaintextBytes: int64(len(first)), PlaintextSHA256: sha256.Sum256(first)},
			OriginalName: "../contract.pdf"},
		{ArchiveEntry: domain.ArchiveEntry{DocumentID: secondID, VersionID: secondVersion,
			Order: 1, PlaintextBytes: int64(len(second)), PlaintextSHA256: sha256.Sum256(second)},
			OriginalName: "proof\r\n.txt"},
	}
	expected, err := zipSize(entries)
	if err != nil {
		t.Fatal(err)
	}
	var output bytes.Buffer
	if err = writeZIP(context.Background(), &output, entries,
		zipFixtureSource{firstVersion: first, secondVersion: second},
		func(int) error { return nil }); err != nil {
		t.Fatal(err)
	}
	if int64(output.Len()) != expected {
		t.Fatalf("zip bytes = %d, want %d", output.Len(), expected)
	}
	reader, err := zip.NewReader(bytes.NewReader(output.Bytes()), int64(output.Len()))
	if err != nil || len(reader.File) != 2 {
		t.Fatalf("read zip: files=%d err=%v", len(reader.File), err)
	}
	if reader.File[0].Name != "00001-_contract.pdf" ||
		reader.File[1].Name != "00002-proof__.txt" {
		t.Fatalf("unsafe names: %q %q", reader.File[0].Name, reader.File[1].Name)
	}
}

func zipFixtureID(t *testing.T, value string) domain.ID {
	t.Helper()
	id, err := domain.ParseID(value)
	if err != nil {
		t.Fatal(err)
	}
	return id
}
