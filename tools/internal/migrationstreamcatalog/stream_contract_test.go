package migrationstreamcatalog

import (
	"strings"
	"testing"
)

func TestLoadRejectsInvalidStreamContracts(t *testing.T) {
	tests := []struct {
		name   string
		mutate func(*Stream)
		want   string
	}{
		{"empty_directory", func(stream *Stream) { stream.Directory = "" }, "directory"},
		{"traversal_directory", func(stream *Stream) { stream.Directory = "../migrations" }, "directory"},
		{"backslash_directory", func(stream *Stream) { stream.Directory = `core\iam` }, "directory"},
		{"colon_directory", func(stream *Stream) { stream.Directory = "C:/iam" }, "directory"},
		{"invalid_scope", func(stream *Stream) { stream.Scope = "global" }, "scope"},
		{"zero_start", func(stream *Stream) { stream.VersionFrom = 0 }, "version range"},
		{"reversed_range", func(stream *Stream) { stream.VersionTo = 0 }, "version range"},
		{"invalid_phase", func(stream *Stream) { stream.Phase = "temporary" }, "phase"},
		{"zero_database_generation", func(stream *Stream) {
			stream.TargetDatabaseGeneration = 0
		}, "generation"},
		{"zero_runtime_generation", func(stream *Stream) {
			stream.NextMinimumRuntimeGeneration = 0
		}, "generation"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			stream := validStream("iam", 10)
			test.mutate(&stream)
			writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
			writeOwner(t, root, ownerDocumentVersion, "iam", []Stream{stream})
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsUnsortedAndDuplicateOwnerStreams(t *testing.T) {
	tests := []struct {
		name    string
		streams []Stream
		want    string
	}{
		{"unsorted", []Stream{validStream("iam", 20), validStream("iam", 10)}, "sorted"},
		{"duplicate", []Stream{validStream("iam", 10), validStream("iam", 10)}, "duplicate"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeRoot(t, root, catalogVersion, []string{streamPath("iam")})
			writeOwner(t, root, ownerDocumentVersion, "iam", test.streams)
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRejectsOverlappingGlobalRanges(t *testing.T) {
	root := t.TempDir()
	alpha := validStream("alpha", 10)
	alpha.VersionTo = 20
	zeta := validStream("zeta", 20)
	zeta.VersionTo = 30
	writeRoot(t, root, catalogVersion, []string{streamPath("alpha"), streamPath("zeta")})
	writeOwner(t, root, ownerDocumentVersion, "alpha", []Stream{alpha})
	writeOwner(t, root, ownerDocumentVersion, "zeta", []Stream{zeta})
	_, err := Load(root)
	if err == nil || !strings.Contains(err.Error(), "overlap") {
		t.Fatalf("overlap error = %v", err)
	}
}
