package profilebackfill

import (
	"crypto/sha256"
	"time"

	"github.com/google/uuid"
)

const RecoveryParentFenced = "fenced"

type RecoveryParent struct {
	ID                  uuid.UUID
	RecoveryRootPlanID  uuid.UUID
	RecoverySequence    int64
	State               string
	PlanSHA256          [sha256.Size]byte
	SourceSchemaID      uuid.UUID
	Target              TargetSchema
	ConverterVersion    string
	Mappings            []Mapping
	ExistingChildPlanID uuid.UUID
}

type RecoveryResource struct {
	ID                   uuid.UUID
	SchemaVersionID      uuid.UUID
	Revision             int64
	SourceSHA256         [sha256.Size]byte
	ExpectedTargetSHA256 [sha256.Size]byte
}

type RecoveryInventory struct {
	parentPlanID          uuid.UUID
	sourceSchemaID        uuid.UUID
	targetSchemaID        uuid.UUID
	inheritedTargetCount  int64
	inheritedTargetSHA256 [sha256.Size]byte
	remainingSourceCount  int64
	pageSize              int
	pages                 []recoveryPageEvidence
	validated             bool
}

func (inventory RecoveryInventory) InheritedTargetCount() int64 {
	return inventory.inheritedTargetCount
}

func (inventory RecoveryInventory) InheritedTargetSHA256() [sha256.Size]byte {
	return inventory.inheritedTargetSHA256
}

func (inventory RecoveryInventory) RemainingSourceCount() int64 {
	return inventory.remainingSourceCount
}

type RecoveryPlanningInput struct {
	ID             uuid.UUID
	Parent         RecoveryParent
	Inventory      RecoveryInventory
	PageSize       int
	ScheduledRunAt time.Time
}

type RecoveryPlan struct {
	id                    uuid.UUID
	recoveryOfPlanID      uuid.UUID
	recoveryRootPlanID    uuid.UUID
	recoverySequence      int64
	parentPlanSHA256      [sha256.Size]byte
	sourceSchemaID        uuid.UUID
	target                TargetSchema
	converterVersion      string
	mappings              []Mapping
	inheritedTargetCount  int64
	inheritedTargetSHA256 [sha256.Size]byte
	remainingSourceCount  int64
	pageSize              int
	initialRollingSHA256  [sha256.Size]byte
	pages                 []Page
	validated             bool
}

func (plan RecoveryPlan) ID() uuid.UUID                 { return plan.id }
func (plan RecoveryPlan) RecoveryOfPlanID() uuid.UUID   { return plan.recoveryOfPlanID }
func (plan RecoveryPlan) RecoveryRootPlanID() uuid.UUID { return plan.recoveryRootPlanID }
func (plan RecoveryPlan) RecoverySequence() int64       { return plan.recoverySequence }
func (plan RecoveryPlan) ParentPlanSHA256() [sha256.Size]byte {
	return plan.parentPlanSHA256
}
func (plan RecoveryPlan) SourceSchemaID() uuid.UUID { return plan.sourceSchemaID }
func (plan RecoveryPlan) Target() TargetSchema      { return cloneTargetSchema(plan.target) }
func (plan RecoveryPlan) ConverterVersion() string  { return plan.converterVersion }
func (plan RecoveryPlan) InheritedTargetCount() int64 {
	return plan.inheritedTargetCount
}
func (plan RecoveryPlan) InheritedTargetSHA256() [sha256.Size]byte {
	return plan.inheritedTargetSHA256
}
func (plan RecoveryPlan) RemainingSourceCount() int64 { return plan.remainingSourceCount }
func (plan RecoveryPlan) PageSize() int               { return plan.pageSize }
func (plan RecoveryPlan) InitialRollingSHA256() [sha256.Size]byte {
	return plan.initialRollingSHA256
}
func (plan RecoveryPlan) Mappings() []Mapping { return cloneMappings(plan.mappings) }
func (plan RecoveryPlan) Pages() []Page       { return append([]Page(nil), plan.pages...) }

type RecoveryCreateCommand struct {
	Plan     RecoveryPlan
	Evidence Evidence
}

type PreparedRecoveryPlan struct {
	Command    RecoveryCreateCommand
	Document   []byte
	PlanSHA256 [sha256.Size]byte
	FirstJob   *FirstJob
}

type RecoveryState string

const (
	RecoveryPlanned   RecoveryState = "planned"
	RecoveryActivated RecoveryState = "activated"
)

type RecoveryResult struct {
	PlanID             uuid.UUID
	Revision           int64
	AuthorizationClock int64
	State              RecoveryState
	Replayed           bool
}
