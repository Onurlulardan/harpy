package migration

import "github.com/Onurlulardan/harpy/internal/migrationbundle"

func SelectFreshDocuments(documents []Document) ([]Document, error) {
	if len(documents) == 0 {
		return nil, errHarpyFreshSchemaAbsent
	}
	sources := make([]migrationbundle.Source, len(documents))
	byPath := make(map[string]Document, len(documents))
	for index, document := range documents {
		sources[index] = migrationbundle.Source{
			Owner: document.Owner(), Scope: document.Scope(),
			Path: document.Path(), Body: document.Content(),
		}
		byPath[document.Path()] = document
	}
	selected, err := migrationbundle.Select(sources)
	if err != nil {
		return nil, err
	}
	documents = make([]Document, len(selected))
	for index, source := range selected {
		documents[index] = byPath[source.Path]
	}
	return documents, nil
}
