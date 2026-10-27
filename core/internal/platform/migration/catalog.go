package migration

import (
	"fmt"
	"io/fs"

	postgresv1 "github.com/Onurlulardan/harpy/contracts/postgres/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

type Catalog struct {
	selected      map[string][]postgresv1.SelectedMigration
	files         map[string]*memoryFS
	supersessions map[string][]HistorySupersession
}

func NewCatalog(
	registry *ReleaseRegistry,
	artifacts attestation.VerifiedPostgresArtifactSet,
	documents []Document,
) (*Catalog, error) {
	if !artifacts.IsVerified() {
		return nil, fmt.Errorf("migration artifacts are not verified")
	}
	if err := registry.Authorize(artifacts.Artifacts().Migration); err != nil {
		return nil, err
	}
	manifest := artifacts.Artifacts().Migration.Value()
	selected := artifacts.SelectedMigrations()
	bound, err := bindDocuments(manifest, artifacts.DatabaseOrigin(), selected, documents)
	if err != nil {
		return nil, err
	}
	files, scopedSelected, err := partitionCatalog(bound, selected)
	if err != nil {
		return nil, err
	}
	supersessions, err := bindHistorySupersessions(selected, documents)
	if err != nil {
		return nil, err
	}
	return &Catalog{
		selected: scopedSelected, files: files, supersessions: supersessions,
	}, nil
}

func (catalog *Catalog) HistorySupersessions(scope string) []HistorySupersession {
	return append([]HistorySupersession(nil), catalog.supersessions[scope]...)
}

func (catalog *Catalog) FS(scope string) (fs.FS, error) {
	files, exists := catalog.files[scope]
	if !exists {
		return nil, fmt.Errorf("migration scope %s is absent from catalog", scope)
	}
	return files, nil
}

func (catalog *Catalog) Selected(scope string) ([]postgresv1.SelectedMigration, error) {
	selected, exists := catalog.selected[scope]
	if !exists {
		return nil, fmt.Errorf("migration scope %s is absent from catalog", scope)
	}
	return append([]postgresv1.SelectedMigration(nil), selected...), nil
}
