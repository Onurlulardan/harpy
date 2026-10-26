package v1adoption

import (
	"context"

	adoption "github.com/Onurlulardan/harpy/core/internal/modules/work/application/v1adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (writer *CanonicalWriter) persistBoardRecord(ctx context.Context,
	record adoption.TargetRecord,
) error {
	switch record.Kind {
	case adoption.TargetBoard:
		var value adoption.TargetBoardDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		creatorID, err := parseID(value.CreatorID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptBoardSQL, record.ID, creatorID, value.Name,
			value.Description, value.Lifecycle, value.CreatedAt, value.UpdatedAt, value.ArchivedAt)
	case adoption.TargetColumn:
		var value adoption.TargetColumnDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		boardID, err := parseID(value.BoardID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptColumnSQL, record.ID, boardID, value.Name,
			value.Color, value.Status, value.Position, value.CreatedAt, value.UpdatedAt)
	case adoption.TargetMembership:
		var value adoption.TargetMembershipDocument
		if err := decodeDocument(record, &value); err != nil {
			return err
		}
		boardID, err := parseID(value.BoardID)
		if err != nil {
			return err
		}
		principalID, err := parseID(value.PrincipalID)
		if err != nil {
			return err
		}
		return writer.executeOne(ctx, adoptMembershipSQL, record.ID, boardID,
			value.PrincipalKind, principalID, value.Role, value.CreatedAt, value.UpdatedAt)
	default:
		return ports.ErrInvalid
	}
}
