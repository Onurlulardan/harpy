package v1adoption

import (
	"context"
	"crypto/sha256"
	"encoding/json"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func sourceRecord(kind adoption.SourceKind, id domain.ID, value any) (adoption.SourceRecord, error) {
	document, err := json.Marshal(value)
	if err != nil || !id.Valid() {
		return adoption.SourceRecord{}, ports.ErrUnavailable
	}
	return adoption.SourceRecord{Kind: kind, ID: id, Document: document,
		SHA256: sha256.Sum256(document)}, nil
}

func queryRows(ctx context.Context, source *LegacySource, statement string) (Rows, error) {
	rows, err := source.database.Query(ctx, statement, source.tenantID)
	if err != nil {
		return nil, ports.ErrUnavailable
	}
	return rows, nil
}
