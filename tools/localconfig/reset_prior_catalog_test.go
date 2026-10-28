package main

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestResetAcceptsIntactPriorCatalogRejectedByEnsure(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	cases := []struct {
		name    string
		count   int
		removed []string
	}{
		{"pre-minio-root", 109, []string{
			"minio-root-access-key.txt", "minio-root-secret-key.txt", "rabbitmq.conf",
		}},
		{"pre-rabbitmq-conf", 111, []string{"rabbitmq.conf"}},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			output := priorCatalogSetForTest(t, now, test.count, test.removed...)
			if _, err := ensureArtifactSet(output, now.Add(time.Hour)); err == nil {
				t.Fatal("normal ensure accepted a prior artifact catalog")
			}
			removed, _, err := resetArtifactSet(output, now.Add(time.Hour))
			if err != nil || !removed {
				t.Fatalf("intact prior catalog reset: removed=%t err=%v", removed, err)
			}
			if _, err := os.Lstat(output); !os.IsNotExist(err) {
				t.Fatalf("reset prior catalog still exists: %v", err)
			}
		})
	}
}

func TestResetRefusesDamagedPriorCatalog(t *testing.T) {
	now := time.Date(2026, 7, 22, 12, 0, 0, 0, time.UTC)
	output := priorCatalogSetForTest(t, now, 111, "rabbitmq.conf")
	path := filepath.Join(output, "s3-access-key-id.txt")
	if err := os.WriteFile(path, []byte("damaged"), 0o600); err != nil {
		t.Fatal(err)
	}
	removed, _, err := resetArtifactSet(output, now.Add(time.Hour))
	if err == nil || removed {
		t.Fatalf("damaged prior catalog reset: removed=%t err=%v", removed, err)
	}
	if _, err := os.Lstat(output); err != nil {
		t.Fatalf("damaged prior catalog was removed: %v", err)
	}
}

func priorCatalogSetForTest(
	t *testing.T,
	now time.Time,
	wantCount int,
	removedNames ...string,
) string {
	t.Helper()
	output := filepath.Join(t.TempDir(), "secrets")
	if _, err := ensureArtifactSet(output, now); err != nil {
		t.Fatal(err)
	}
	removed := make(map[string]bool, len(removedNames))
	for _, name := range removedNames {
		removed[name] = true
	}
	for name := range removed {
		if err := os.Remove(filepath.Join(output, name)); err != nil {
			t.Fatal(err)
		}
	}
	manifestPath := filepath.Join(output, manifestFileName)
	content, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	manifest, err := decodeSetManifest(content)
	if err != nil {
		t.Fatal(err)
	}
	files := manifest.Files[:0]
	for _, file := range manifest.Files {
		if !removed[file.Name] {
			files = append(files, file)
		}
	}
	if len(files) != wantCount {
		t.Fatalf("prior manifest file count = %d, want %d", len(files), wantCount)
	}
	manifest.Files = files
	manifest.ArtifactContractRevision = 0
	content, err = marshalJSON(manifest)
	if err != nil {
		t.Fatalf("write prior manifest: %v", err)
	}
	if err := os.WriteFile(manifestPath, content, 0o600); err != nil {
		t.Fatalf("write prior manifest: %v", err)
	}
	return output
}
