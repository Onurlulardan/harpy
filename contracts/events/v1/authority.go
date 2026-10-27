package eventsv1

import "fmt"

type ScopeKind string

const (
	ScopeTenant ScopeKind = "tenant"
	ScopeGlobal ScopeKind = "global"
)

type WriterOrigin string

const (
	WriterV1 WriterOrigin = "v1"
	WriterV2 WriterOrigin = "v2"
)

type Authority struct {
	ScopeKind         ScopeKind    `json:"scope_kind"`
	ScopeID           string       `json:"scope_id"`
	Domain            string       `json:"domain"`
	CutoverGeneration int64        `json:"cutover_generation"`
	WriterOrigin      WriterOrigin `json:"writer_origin"`
	GroupID           string       `json:"group_id"`
	GroupGeneration   int64        `json:"group_generation"`
}

func (value Authority) Validate() error {
	if value.ScopeKind != ScopeTenant && value.ScopeKind != ScopeGlobal {
		return fmt.Errorf("authority scope kind is invalid")
	}
	if !validUUID(value.ScopeID) || !validDomain(value.Domain) ||
		value.CutoverGeneration < 1 {
		return fmt.Errorf("authority identity is invalid")
	}
	if value.WriterOrigin != WriterV1 && value.WriterOrigin != WriterV2 {
		return fmt.Errorf("writer origin is invalid")
	}
	if !validDomain(value.GroupID) || value.GroupGeneration < 1 {
		return fmt.Errorf("authority group stamp is invalid")
	}
	return nil
}

func (value Authority) ValidateTenant(tenantID *string) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if value.ScopeKind == ScopeTenant && (tenantID == nil || *tenantID != value.ScopeID) {
		return fmt.Errorf("tenant authority must match tenant ID")
	}
	if value.ScopeKind == ScopeGlobal && tenantID != nil {
		return fmt.Errorf("global authority cannot carry tenant ID")
	}
	return nil
}

func (value Authority) ValidateConsumer(tenantID *string) error {
	if err := value.ValidateTenant(tenantID); err != nil {
		return err
	}
	if value.WriterOrigin != WriterV2 {
		return fmt.Errorf("consumer authority writer origin must be v2")
	}
	return nil
}
