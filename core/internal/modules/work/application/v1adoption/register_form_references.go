package v1adoption

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
	"github.com/Onurlulardan/harpy/core/internal/modules/work/ports"
)

func (service *Service) registerFormReferences(ctx context.Context,
	scope ports.Scope, plan *Plan,
) error {
	if plan == nil || len(plan.Issues) != 0 {
		return nil
	}
	occurredAt, err := adoptionOperationTime(plan.RunID)
	if err != nil {
		return err
	}
	for index := range plan.Records {
		if plan.Records[index].Kind != TargetFormLink {
			continue
		}
		if err = service.registerFormReference(
			ctx, scope, &plan.Records[index], plan.RunID, occurredAt); err != nil {
			return err
		}
	}
	return nil
}

func (service *Service) registerFormReference(ctx context.Context,
	scope ports.Scope, record *TargetRecord, operationID domain.ID, occurredAt time.Time,
) error {
	var document TargetFormLinkDocument
	if record == nil || json.Unmarshal(record.Document, &document) != nil ||
		document.ReferenceEpoch != 0 || document.PermissionRevision < 1 ||
		(document.ResourceKind != "TASK" && document.ResourceKind != "CALENDAR_EVENT") {
		return ports.ErrInvalid
	}
	resourceID, resourceErr := domain.ParseID(document.ResourceID)
	formID, formErr := domain.ParseID(document.FormID)
	versionID, versionErr := domain.ParseID(document.VersionID)
	if resourceErr != nil || formErr != nil || versionErr != nil {
		return ports.ErrInvalid
	}
	command := ports.ResourceReferenceCommand{Link: domain.ResourceLink{
		ID: record.ID, WorkResourceID: resourceID,
		TargetID: formID, TargetVersionID: versionID,
		TargetOwner: domain.SourceForms, TargetKind: "FORM",
		PermissionRevision: document.PermissionRevision, RowVersion: 1,
	}, OperationID: operationID, OccurredAt: occurredAt}
	receipt, err := service.dependencies.Forms.RegisterRead(ctx, scope, command)
	if err != nil || !receipt.Matches(command) {
		if err != nil {
			return err
		}
		return ports.ErrUnavailable
	}
	document.ReferenceEpoch = receipt.ReferenceEpoch
	record.Document, err = json.Marshal(document)
	if err != nil {
		return ports.ErrUnavailable
	}
	record.SHA256 = sha256.Sum256(record.Document)
	return nil
}

func adoptionOperationTime(id domain.ID) (time.Time, error) {
	if !id.UUIDv7() {
		return time.Time{}, ports.ErrInvalid
	}
	milliseconds := int64(id[0])<<40 | int64(id[1])<<32 |
		int64(id[2])<<24 | int64(id[3])<<16 | int64(id[4])<<8 | int64(id[5])
	return time.UnixMilli(milliseconds).UTC(), nil
}
