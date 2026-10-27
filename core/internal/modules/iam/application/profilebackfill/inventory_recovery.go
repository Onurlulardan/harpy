package profilebackfill

import (
	"bytes"
	"crypto/sha256"

	"github.com/google/uuid"
)

type RecoveryInventoryAccumulator struct {
	parentPlanID          uuid.UUID
	sourceSchemaID        uuid.UUID
	targetSchemaID        uuid.UUID
	pageSize              int
	inheritedTargetCount  int64
	inheritedTargetSHA256 [sha256.Size]byte
	remainingSourceCount  int64
	previousResourceID    uuid.UUID
	previousSourceID      uuid.UUID
	current               []RecoveryResource
	pages                 []recoveryPageEvidence
	finished              bool
}

func NewRecoveryInventoryAccumulator(
	parentPlanID uuid.UUID,
	sourceSchemaID uuid.UUID,
	targetSchemaID uuid.UUID,
	pageSize int,
) (*RecoveryInventoryAccumulator, error) {
	if !stableID(parentPlanID) || !stableID(sourceSchemaID) ||
		!stableID(targetSchemaID) || sourceSchemaID == targetSchemaID ||
		pageSize < 1 || pageSize > MaximumPageSize {
		return nil, ErrInvalid
	}
	return &RecoveryInventoryAccumulator{
		parentPlanID: parentPlanID, sourceSchemaID: sourceSchemaID,
		targetSchemaID: targetSchemaID, pageSize: pageSize,
		inheritedTargetSHA256: sha256.Sum256(nil),
		current:               make([]RecoveryResource, 0, pageSize),
	}, nil
}

func (accumulator *RecoveryInventoryAccumulator) AppendTarget(
	resource SourceResource,
) error {
	if accumulator == nil || accumulator.finished || resource.Validate() != nil ||
		resource.SchemaVersionID != accumulator.targetSchemaID ||
		accumulator.advance(resource.ID) != nil {
		return ErrInvalid
	}
	accumulator.inheritedTargetSHA256 = rollRecoveryTargetDigest(
		accumulator.inheritedTargetSHA256, resource,
	)
	accumulator.inheritedTargetCount++
	return nil
}

func (accumulator *RecoveryInventoryAccumulator) AppendRemaining(
	resource RecoveryResource,
) error {
	if accumulator == nil || accumulator.finished || !validRecoveryResource(resource) ||
		resource.SchemaVersionID == accumulator.targetSchemaID ||
		accumulator.advance(resource.ID) != nil {
		return ErrInvalid
	}
	accumulator.current = append(accumulator.current, resource)
	accumulator.remainingSourceCount++
	if len(accumulator.current) == accumulator.pageSize {
		return accumulator.flush()
	}
	return nil
}

func (accumulator *RecoveryInventoryAccumulator) Seal() (RecoveryInventory, error) {
	if accumulator == nil || accumulator.finished {
		return RecoveryInventory{}, ErrInvalid
	}
	if err := accumulator.flush(); err != nil {
		return RecoveryInventory{}, err
	}
	accumulator.finished = true
	pages := append([]recoveryPageEvidence(nil), accumulator.pages...)
	clear(accumulator.pages)
	accumulator.pages = nil
	return RecoveryInventory{
		parentPlanID:          accumulator.parentPlanID,
		sourceSchemaID:        accumulator.sourceSchemaID,
		targetSchemaID:        accumulator.targetSchemaID,
		inheritedTargetCount:  accumulator.inheritedTargetCount,
		inheritedTargetSHA256: accumulator.inheritedTargetSHA256,
		remainingSourceCount:  accumulator.remainingSourceCount,
		pageSize:              accumulator.pageSize,
		pages:                 pages,
		validated:             true,
	}, nil
}

func (accumulator *RecoveryInventoryAccumulator) advance(id uuid.UUID) error {
	if !stableID(id) || accumulator.previousResourceID != uuid.Nil &&
		bytes.Compare(accumulator.previousResourceID[:], id[:]) >= 0 {
		return ErrInvalid
	}
	accumulator.previousResourceID = id
	return nil
}

func (accumulator *RecoveryInventoryAccumulator) flush() error {
	if len(accumulator.current) == 0 {
		return nil
	}
	page, err := newRecoveryPageEvidence(
		accumulator.previousSourceID, accumulator.current,
	)
	if err != nil {
		return err
	}
	accumulator.pages = append(accumulator.pages, page)
	accumulator.previousSourceID = page.throughResourceID
	clear(accumulator.current)
	accumulator.current = accumulator.current[:0]
	return nil
}

func validRecoveryResource(resource RecoveryResource) bool {
	return stableID(resource.ID) && stableID(resource.SchemaVersionID) &&
		resource.Revision > 0 && resource.SourceSHA256 != [sha256.Size]byte{} &&
		resource.ExpectedTargetSHA256 != [sha256.Size]byte{}
}
