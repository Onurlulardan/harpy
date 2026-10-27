package application

import (
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/domain"
	"github.com/google/uuid"
)

const (
	RetentionScopeKind       = "global"
	RetentionAuthorityGroup  = "usage"
	RetentionAuthorityDomain = "usage"
	RetentionWriterOrigin    = "v2"
)

type RetentionAuthorityInput struct {
	EnvironmentID    uuid.UUID
	ScopeKind        string
	ScopeID          uuid.UUID
	GroupID          string
	GroupGeneration  int64
	Domain           string
	DomainGeneration int64
	WriterOrigin     string
}

type RetentionAuthority struct {
	input  RetentionAuthorityInput
	sealed bool
}

// NewRetentionAuthority seals evidence already verified by the global job
// transaction owner; positive generations alone are not a live DB check.
func NewRetentionAuthority(input RetentionAuthorityInput) (RetentionAuthority, error) {
	value := RetentionAuthority{input: input, sealed: true}
	if !value.validFor(input.EnvironmentID) {
		return RetentionAuthority{}, domain.ErrForbidden
	}
	return value, nil
}

func (authority RetentionAuthority) validFor(environmentID uuid.UUID) bool {
	input := authority.input
	return authority.sealed && environmentID != uuid.Nil &&
		input.EnvironmentID == environmentID &&
		input.ScopeKind == RetentionScopeKind && input.ScopeID == environmentID &&
		input.GroupID == RetentionAuthorityGroup && input.GroupGeneration > 0 &&
		input.Domain == RetentionAuthorityDomain && input.DomainGeneration > 0 &&
		input.WriterOrigin == RetentionWriterOrigin
}

func (authority RetentionAuthority) EnvironmentID() uuid.UUID {
	if !authority.validFor(authority.input.EnvironmentID) {
		return uuid.Nil
	}
	return authority.input.EnvironmentID
}

func (authority RetentionAuthority) GroupGeneration() int64 {
	if !authority.validFor(authority.input.EnvironmentID) {
		return 0
	}
	return authority.input.GroupGeneration
}

func (authority RetentionAuthority) DomainGeneration() int64 {
	if !authority.validFor(authority.input.EnvironmentID) {
		return 0
	}
	return authority.input.DomainGeneration
}
