package asyncadoptionv1

import "fmt"

type Item struct {
	ContractVersion        string         `json:"contract_version"`
	InventoryClass         InventoryClass `json:"inventory_class"`
	SourceSurface          SourceSurface  `json:"source_surface"`
	SourceName             string         `json:"source_name"`
	SourceItemID           string         `json:"source_item_id"`
	SourceContractVersion  string         `json:"source_contract_version"`
	SourceRevision         int            `json:"source_revision"`
	SourcePayloadSHA256    string         `json:"source_payload_sha256"`
	TransformVersion       string         `json:"transform_version"`
	IdempotencyIdentity    string         `json:"idempotency_identity"`
	TargetKind             string         `json:"target_kind"`
	TargetPayloadRevision  int            `json:"target_payload_revision"`
	TargetPayloadReference string         `json:"target_payload_reference"`
	TargetPayloadSHA256    string         `json:"target_payload_sha256"`
	TargetRunAt            string         `json:"target_run_at"`
	TargetPriority         int32          `json:"target_priority"`
	TargetMaximumAttempts  int            `json:"target_maximum_attempts"`
	TargetEffectClass      EffectClass    `json:"target_effect_class"`
	TargetEffectKey        *string        `json:"target_effect_idempotency_key"`
	CapturedAt             string         `json:"captured_at"`
	ItemSHA256             string         `json:"item_sha256"`
}

func NewItem(value Item) (Item, error) {
	value.ContractVersion = ItemContractVersion
	value.TargetEffectKey = cloneString(value.TargetEffectKey)
	value.ItemSHA256 = ""
	digest, err := value.digest()
	if err != nil {
		return Item{}, err
	}
	value.ItemSHA256 = digest
	return value, value.Validate()
}

func (value Item) Validate() error {
	if value.ContractVersion != ItemContractVersion || !validInventoryPair(value.InventoryClass, value.SourceSurface) {
		return fmt.Errorf("async adoption item source class is invalid")
	}
	if !boundedName.MatchString(value.SourceName) || !validText(value.SourceItemID, MaximumIdentityBytes) ||
		!boundedName.MatchString(value.SourceContractVersion) || value.SourceRevision < 1 {
		return fmt.Errorf("async adoption item source identity is invalid")
	}
	if !lowerSHA256.MatchString(value.SourcePayloadSHA256) || !boundedName.MatchString(value.TransformVersion) ||
		!validText(value.IdempotencyIdentity, MaximumIdentityBytes) {
		return fmt.Errorf("async adoption item source evidence is invalid")
	}
	if !boundedName.MatchString(value.TargetKind) || value.TargetPayloadRevision < 1 ||
		!validText(value.TargetPayloadReference, MaximumReferenceBytes) ||
		!lowerSHA256.MatchString(value.TargetPayloadSHA256) || !validTimestamp(value.TargetRunAt) ||
		value.TargetMaximumAttempts < 1 || value.TargetMaximumAttempts > 25 {
		return fmt.Errorf("async adoption target job is invalid")
	}
	if value.TargetEffectClass == EffectExternal {
		if value.TargetEffectKey == nil || *value.TargetEffectKey != value.IdempotencyIdentity {
			return fmt.Errorf("external adoption identity is invalid")
		}
	} else if value.TargetEffectClass != EffectDatabaseTransactional || value.TargetEffectKey != nil {
		return fmt.Errorf("database adoption identity is invalid")
	}
	if !validTimestamp(value.CapturedAt) || !lowerSHA256.MatchString(value.ItemSHA256) {
		return fmt.Errorf("async adoption item capture evidence is invalid")
	}
	expected, err := value.digest()
	if err != nil || expected != value.ItemSHA256 {
		return fmt.Errorf("async adoption item checksum differs")
	}
	return validateKnownV1Surface(value)
}

func (value Item) digest() (string, error) {
	type digestWire Item
	wire := digestWire(value)
	wire.ItemSHA256 = ""
	return canonicalDigest(wire)
}

func validInventoryPair(class InventoryClass, surface SourceSurface) bool {
	return class == ClassQueue && surface == SurfaceQueue ||
		class == ClassTimer && (surface == SurfaceScheduler || surface == SurfaceDelayedPlugin || surface == SurfaceTTLQueue) ||
		class == ClassFuture && surface == SurfaceFutureState ||
		class == ClassRunning && surface == SurfaceRunningState
}

func validateKnownV1Surface(value Item) error {
	want := map[string]struct {
		class    InventoryClass
		surface  SourceSurface
		contract string
	}{
		V1ScheduleQueue: {ClassQueue, SurfaceQueue, V1QueueContract},
	}
	policy, known := want[value.SourceName]
	if known && (value.InventoryClass != policy.class || value.SourceSurface != policy.surface ||
		value.SourceContractVersion != policy.contract) {
		return fmt.Errorf("known V1 async surface semantics differ")
	}
	return nil
}
