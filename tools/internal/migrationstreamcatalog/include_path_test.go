package migrationstreamcatalog

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadRejectsInvalidIncludePaths(t *testing.T) {
	paths := []string{
		"../iam.json",
		"/contracts/storage/migration-streams/iam.json",
		`contracts\storage\migration-streams\iam.json`,
		"C:/contracts/storage/migration-streams/iam.json",
		"contracts/storage/migration-streams/nested/iam.json",
		"contracts/storage/migration-streams/Iam.json",
		"contracts/storage/migration-streams/iam-owner.json",
		"contracts/storage/migration-streams/iam.yaml",
		"contracts/storage/iam.json",
	}
	for _, include := range paths {
		t.Run(strings.ReplaceAll(include, "/", "_"), func(t *testing.T) {
			root := t.TempDir()
			writeRoot(t, root, catalogVersion, []string{include})
			if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "invalid") {
				t.Fatalf("invalid include %q error = %v", include, err)
			}
		})
	}
}

func TestLoadRejectsUnsortedAndDuplicateIncludes(t *testing.T) {
	tests := []struct {
		name     string
		includes []string
		want     string
	}{
		{"unsorted", []string{streamPath("zeta"), streamPath("alpha")}, "sorted"},
		{"duplicate", []string{streamPath("alpha"), streamPath("alpha")}, "duplicate"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			root := t.TempDir()
			writeRoot(t, root, catalogVersion, test.includes)
			_, err := Load(root)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("Load error = %v, want %q", err, test.want)
			}
		})
	}
}

func TestLoadRequiresAbsoluteCleanRepositoryRoot(t *testing.T) {
	if _, err := Load("."); err == nil || !strings.Contains(err.Error(), "absolute") {
		t.Fatalf("relative root error = %v", err)
	}
	root := t.TempDir() + string(filepath.Separator) + "."
	if _, err := Load(root); err == nil || !strings.Contains(err.Error(), "clean") {
		t.Fatalf("unclean root error = %v", err)
	}
}
