package migrationmanifest

import (
	"fmt"
	"io/fs"
	"os"
	"path"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Onurlulardan/harpy/internal/attestation/migrationsupersession"
)

func Collect(repositoryRoot string, spec SourceSpec) (SourceGroup, error) {
	if !filepath.IsAbs(repositoryRoot) || !fs.ValidPath(spec.Directory) {
		return SourceGroup{}, fmt.Errorf("migration repository root or directory is invalid")
	}
	source := os.DirFS(repositoryRoot)
	var documents []SourceDocument
	err := fs.WalkDir(source, spec.Directory, func(
		name string,
		entry fs.DirEntry,
		walkErr error,
	) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if name == path.Join(spec.Directory, migrationsupersession.CatalogFileName) {
			return nil
		}
		if entry.Type()&fs.ModeSymlink != 0 {
			return fmt.Errorf("migration source %s is a symbolic link", name)
		}
		info, err := entry.Info()
		if err != nil {
			return err
		}
		if !info.Mode().IsRegular() || !strings.HasSuffix(name, ".sql") {
			return fmt.Errorf("migration source %s is not a regular SQL file", name)
		}
		content, err := fs.ReadFile(source, name)
		if err != nil {
			return err
		}
		document, err := NewSourceDocument(spec.Owner, name, content)
		if err != nil {
			return fmt.Errorf("collect migration source %s: %w", name, err)
		}
		documents = append(documents, document)
		return nil
	})
	if err != nil {
		return SourceGroup{}, err
	}
	documents, err = applySourceSupersessions(repositoryRoot, spec, documents)
	if err != nil {
		return SourceGroup{}, err
	}
	sort.Slice(documents, func(left, right int) bool {
		return documents[left].Path() < documents[right].Path()
	})
	return SourceGroup{
		Owner: spec.Owner, Scope: spec.Scope, Phase: spec.Phase,
		TargetDatabaseGeneration:     spec.TargetDatabaseGeneration,
		NextMinimumRuntimeGeneration: spec.NextMinimumRuntimeGeneration,
		VersionFrom:                  spec.VersionFrom,
		VersionTo:                    spec.VersionTo,
		Directory:                    spec.Directory,
		Documents:                    documents,
	}, nil
}
