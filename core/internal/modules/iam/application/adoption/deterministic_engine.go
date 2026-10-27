package adoption

import (
	"fmt"
	"sort"

	"github.com/google/uuid"
)

type Compiler struct {
	normalizers normalizers
}

func NewCompiler(email, portal NormalizeIdentifier) (*Compiler, error) {
	values, err := newNormalizers(email, portal)
	if err != nil {
		return nil, err
	}
	return &Compiler{normalizers: values}, nil
}

func (compiler *Compiler) Evaluate(snapshot Snapshot, mapping Mapping) (Plan, error) {
	if compiler == nil {
		return Plan{}, fmt.Errorf("IAM V1 adoption compiler is required")
	}
	if snapshot.TenantID == uuid.Nil || mapping.TenantID != snapshot.TenantID ||
		mapping.ContractVersion != MappingContractVersion {
		return Plan{}, fmt.Errorf("IAM V1 adoption scope or mapping contract is invalid")
	}
	state := newValidationState(snapshot, mapping, compiler.normalizers)
	state.validateOrganizations()
	state.validateRoles()
	state.validateUsers()
	state.validatePermissions()
	state.validateExtensions()
	state.validateProtectedAdministrator()
	state.sortIssues()
	categories, err := buildCategoryEvidence(state.snapshot)
	if err != nil {
		return Plan{}, err
	}
	mappingSHA, err := canonicalSHA256(normalizedMappingForHash(state.mapping))
	if err != nil {
		return Plan{}, fmt.Errorf("canonicalize IAM V1 adoption mapping: %w", err)
	}
	sourceSHA, err := canonicalSHA256(struct {
		TenantID   string             `json:"tenant_id"`
		Categories []CategoryEvidence `json:"categories"`
	}{TenantID: snapshot.TenantID.String(), Categories: categories})
	if err != nil {
		return Plan{}, fmt.Errorf("canonicalize IAM V1 adoption source: %w", err)
	}
	return Plan{
		ContractVersion: PlanContractVersion,
		TenantID:        snapshot.TenantID.String(),
		MappingSHA256:   mappingSHA,
		SourceSHA256:    sourceSHA,
		Categories:      categories,
		Issues:          state.issues,
		Eligible:        len(state.issues) == 0,
	}, nil
}

type validationState struct {
	snapshot    Snapshot
	mapping     Mapping
	issues      []Issue
	orgs        map[uuid.UUID]SourceOrganization
	roles       map[uuid.UUID]SourceRole
	users       map[uuid.UUID]SourceUser
	resources   map[uuid.UUID]SourceResource
	actions     map[uuid.UUID]SourceAction
	normalizers normalizers
}

func newValidationState(snapshot Snapshot, mapping Mapping, values normalizers) *validationState {
	return &validationState{
		snapshot: snapshot, mapping: mapping, issues: make([]Issue, 0),
		orgs: make(map[uuid.UUID]SourceOrganization), roles: make(map[uuid.UUID]SourceRole),
		users: make(map[uuid.UUID]SourceUser), resources: make(map[uuid.UUID]SourceResource),
		actions: make(map[uuid.UUID]SourceAction), normalizers: values,
	}
}

func (state *validationState) add(value Issue) {
	state.issues = append(state.issues, value)
}

func (state *validationState) sortIssues() {
	sort.Slice(state.issues, func(left, right int) bool {
		a, b := state.issues[left], state.issues[right]
		if a.Category != b.Category {
			return a.Category < b.Category
		}
		if a.Code != b.Code {
			return a.Code < b.Code
		}
		return a.SourceKey < b.SourceKey
	})
}
