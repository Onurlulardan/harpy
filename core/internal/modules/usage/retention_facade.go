package usage

import (
	"fmt"

	usagepostgres "github.com/Onurlulardan/harpy/core/internal/modules/usage/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/usage/application"
)

type RetentionAuthorityInput = application.RetentionAuthorityInput
type RetentionAuthority = application.RetentionAuthority
type RetentionJobClaim = application.RetentionJobClaim
type RetentionOutcome = application.RetentionOutcome
type RetentionSuccessor = application.RetentionSuccessor
type RetentionCommands = application.RetentionCommands

const (
	RetentionJobKind         = application.RetentionJobKind
	RetentionPayloadRevision = application.RetentionPayloadRevision
	RetentionBatchLimit      = application.RetentionBatchLimit
	RetentionScopeKind       = application.RetentionScopeKind
	RetentionAuthorityGroup  = application.RetentionAuthorityGroup
	RetentionAuthorityDomain = application.RetentionAuthorityDomain
	RetentionWriterOrigin    = application.RetentionWriterOrigin
	RetentionAge             = application.RetentionAge
	RetentionCycleInterval   = application.RetentionCycleInterval
)

var NewRetentionAuthority = application.NewRetentionAuthority
var RetentionReference = application.RetentionReference

type RetentionJobBinding struct{ ids IDGenerator }

func NewRetentionJobBinding(ids IDGenerator) (RetentionJobBinding, error) {
	if ids == nil {
		return RetentionJobBinding{}, fmt.Errorf("usage retention ID generator is required")
	}
	return RetentionJobBinding{ids: ids}, nil
}

func (binding RetentionJobBinding) Owner() string {
	if binding.ids == nil {
		return ""
	}
	return RetentionAuthorityDomain
}

func (binding RetentionJobBinding) Bind(
	database PostgresDatabase,
) (*RetentionCommands, error) {
	if binding.Owner() == "" {
		return nil, fmt.Errorf("usage retention binding is unavailable")
	}
	store, err := usagepostgres.NewRetentionStore(database)
	if err != nil {
		return nil, fmt.Errorf("bind usage retention owner store: %w", err)
	}
	commands, err := application.NewRetentionCommands(store, binding.ids)
	if err != nil {
		return nil, fmt.Errorf("bind usage retention owner command: %w", err)
	}
	return commands, nil
}
