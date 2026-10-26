package migration

import (
	"fmt"
	"io/fs"
	"path"
	"strings"
)

func CollectEmbedded(
	owner string,
	scope string,
	repositoryRoot string,
	embedded fs.FS,
) ([]Document, error) {
	if !fs.ValidPath(repositoryRoot) || strings.HasSuffix(repositoryRoot, "/") {
		return nil, fmt.Errorf("migration repository root is invalid")
	}
	var documents []Document
	err := fs.WalkDir(embedded, ".", func(name string, entry fs.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() {
			return nil
		}
		if !strings.HasSuffix(name, ".sql") {
			return fmt.Errorf("embedded migration %s is not SQL", name)
		}
		content, err := fs.ReadFile(embedded, name)
		if err != nil {
			return err
		}
		document, err := NewDocument(owner, scope, path.Join(repositoryRoot, name), content)
		if err != nil {
			return fmt.Errorf("collect embedded migration %s: %w", name, err)
		}
		if _, err := InspectDocument(document); err != nil {
			return fmt.Errorf("inspect embedded migration %s: %w", name, err)
		}
		documents = append(documents, document)
		return nil
	})
	if err != nil {
		return nil, err
	}
	if len(documents) == 0 {
		return nil, fmt.Errorf("embedded migration set is empty")
	}
	return documents, nil
}
