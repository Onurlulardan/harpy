package lifecycle

import (
	"context"
	_ "embed"

	catalogapp "github.com/Onurlulardan/harpy/core/internal/modules/catalog/application/lifecycle"
	"github.com/google/uuid"
)

//go:embed sql/tenant/list_demo_components.sql
var listDemoComponentsSQL string

func (storage *store) DemoComponents(
	ctx context.Context,
	installationID uuid.UUID,
) ([]catalogapp.DemoComponent, error) {
	if installationID.Version() != 7 {
		return nil, catalogapp.ErrInvalid
	}
	if !storage.usable(ctx, false) {
		return nil, unavailable()
	}
	rows, err := storage.query(ctx, listDemoComponentsSQL, installationID)
	if err != nil {
		return nil, unavailable()
	}
	defer rows.Close()
	result := make([]catalogapp.DemoComponent, 0, 1)
	for rows.Next() {
		value, scanErr := scanDemoComponent(rows)
		if scanErr != nil || value.InstallationID != installationID {
			return nil, unavailable()
		}
		result = append(result, value)
	}
	if rows.Err() != nil {
		return nil, unavailable()
	}
	return result, nil
}

func scanDemoComponent(source interface{ Scan(...any) error }) (
	catalogapp.DemoComponent,
	error,
) {
	var value catalogapp.DemoComponent
	var recordIDs []byte
	var revision int64
	err := source.Scan(
		&value.InstallationID, &value.PackageID, &value.ComponentID,
		&value.ReleaseID, &value.SourceRunID, &value.PackageSlug,
		&value.PackageVersion, &value.ContentSHA256, &value.State,
		&value.Document, &recordIDs, &revision, &value.UpdatedAt,
	)
	if err != nil {
		return catalogapp.DemoComponent{}, err
	}
	value.UpdatedAt = value.UpdatedAt.UTC()
	value.RecordIDs, err = decodeDemoRecordIDs(recordIDs)
	if err != nil || value.InstallationID.Version() != 7 || value.PackageID == uuid.Nil ||
		value.ComponentID.Version() != 7 || value.ReleaseID.Version() != 7 ||
		value.SourceRunID.Version() != 7 || value.PackageSlug == "" ||
		value.PackageVersion == "" || !digestPattern.MatchString(value.ContentSHA256) ||
		(value.State != catalogapp.DemoStateInstalled && value.State != catalogapp.DemoStateAbsent) ||
		len(value.Document) == 0 || revision < 1 || !canonicalDemoTime(value.UpdatedAt) {
		return catalogapp.DemoComponent{}, unavailable()
	}
	value.Revision = uint64(revision)
	return value, nil
}

var _ catalogapp.DemoStore = (*store)(nil)
