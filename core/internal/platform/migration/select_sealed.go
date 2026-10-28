package migration

import "fmt"

func selectSealedDocuments(
	target map[string]expectedDocument,
	documents []Document,
) ([]Document, error) {
	selected := make([]Document, 0, len(target))
	seen := make(map[string]bool, len(documents))
	for _, document := range documents {
		if seen[document.path] {
			return nil, fmt.Errorf("embedded migration source is duplicated")
		}
		seen[document.path] = true
		expected, required := target[document.path]
		if !required {
			continue
		}
		metadata, err := InspectDocument(document)
		if err != nil {
			return nil, err
		}
		actual := expectedDocument{
			owner: document.owner, scope: document.scope,
			transactionMode: metadata.TransactionMode, sha256: metadata.SHA256,
		}
		if actual != expected {
			return nil, fmt.Errorf(
				"embedded migration source %s is not release-authorized", document.path,
			)
		}
		selected = append(selected, document)
	}
	if len(selected) != len(target) {
		return nil, fmt.Errorf("release migration source set is incomplete")
	}
	return selected, nil
}
