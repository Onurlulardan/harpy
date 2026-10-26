package v1adoption

import (
	"context"
	"encoding/json"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

type CanonicalWriter struct {
	database DatabaseAccess
	tenantID domain.ID
}

func NewCanonicalWriter(database DatabaseAccess, tenantID domain.ID) (*CanonicalWriter, error) {
	if database.Execute == nil || database.CurrentSchema == nil || !tenantID.Valid() {
		return nil, ports.ErrInvalid
	}
	return &CanonicalWriter{database: database, tenantID: tenantID}, nil
}

func (writer *CanonicalWriter) Persist(ctx context.Context, record adoption.TargetRecord) error {
	if writer == nil || ctx == nil || ctx.Err() != nil || !record.ID.Valid() ||
		len(record.Document) == 0 {
		return ports.ErrInvalid
	}
	schema, err := writer.database.CurrentSchema(ctx)
	if err != nil || !schemaPattern.MatchString(schema) || schema == "public" {
		return ports.ErrUnavailable
	}
	switch record.Kind {
	case adoption.TargetBoard, adoption.TargetColumn, adoption.TargetMembership:
		return writer.persistBoardRecord(ctx, record)
	case adoption.TargetTask, adoption.TargetAssignee, adoption.TargetLabel,
		adoption.TargetTaskLabel, adoption.TargetChecklist, adoption.TargetFormLink:
		return writer.persistTaskRecord(ctx, record)
	case adoption.TargetComment, adoption.TargetCalendar, adoption.TargetRecurrence,
		adoption.TargetProjection:
		return writer.persistCollaborationRecord(ctx, record)
	default:
		return ports.ErrInvalid
	}
}

func (writer *CanonicalWriter) Revert(_ context.Context, _ adoption.ReverseRequest,
	evidence adoption.Evidence,
) error {
	if evidence.V2WriteCount != 0 {
		return adoption.ErrBlocked
	}
	return nil
}

func decodeDocument(record adoption.TargetRecord, target any) error {
	if json.Unmarshal(record.Document, target) != nil {
		return ports.ErrUnavailable
	}
	return nil
}

func (writer *CanonicalWriter) executeOne(ctx context.Context, sql string, values ...any) error {
	affected, err := writer.database.Execute(ctx, sql, values...)
	if err != nil {
		return ports.ErrUnavailable
	}
	if affected != 1 {
		return ports.ErrConflict
	}
	return nil
}

func parseID(value string) (domain.ID, error) {
	id, err := domain.ParseID(value)
	if err != nil {
		return domain.ID{}, ports.ErrInvalid
	}
	return id, nil
}

var _ RecordWriter = (*CanonicalWriter)(nil)
