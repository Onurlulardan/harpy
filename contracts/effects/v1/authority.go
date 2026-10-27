package effectsv1

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

// Authority is a validated snapshot of a sealed runtime authority stamp.
// This wire value does not mint or grant authority.
type Authority struct {
	ScopeKind         ScopeKind    `json:"scope_kind"`
	ScopeID           string       `json:"scope_id"`
	GroupID           string       `json:"group_id"`
	GroupGeneration   int64        `json:"group_generation"`
	Domain            string       `json:"domain"`
	CutoverGeneration int64        `json:"cutover_generation"`
	WriterOrigin      WriterOrigin `json:"writer_origin"`
}

func (value Authority) Validate() error {
	if value.ScopeKind != ScopeTenant && value.ScopeKind != ScopeGlobal {
		return fmt.Errorf("authority scope kind is invalid")
	}
	if !validUUID(value.ScopeID) {
		return fmt.Errorf("authority scope ID is invalid")
	}
	if !sqlName.MatchString(value.GroupID) || value.GroupGeneration < 1 {
		return fmt.Errorf("authority group is invalid")
	}
	if !sqlName.MatchString(value.Domain) {
		return fmt.Errorf("authority domain is invalid")
	}
	if value.CutoverGeneration < 1 {
		return fmt.Errorf("cutover generation must be positive")
	}
	if value.WriterOrigin != WriterV1 && value.WriterOrigin != WriterV2 {
		return fmt.Errorf("writer origin is invalid")
	}
	return nil
}

func (value Authority) ValidateTenant(tenantID *string) error {
	if err := value.Validate(); err != nil {
		return err
	}
	if value.ScopeKind == ScopeTenant {
		if tenantID == nil || !validUUID(*tenantID) || *tenantID != value.ScopeID {
			return fmt.Errorf("tenant authority must match tenant ID")
		}
		return nil
	}
	if tenantID != nil {
		return fmt.Errorf("global authority cannot carry tenant ID")
	}
	return nil
}
