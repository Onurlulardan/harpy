package iam

import (
	adoptionpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/adoption"
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profile"
	v1authoritypostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/v1authority"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	platformcutover "github.com/Onurlulardan/harpy/core/internal/platform/cutover"
)

func V1ProtectedAuthorityAdoptionContract() (platformcutover.ApplyContract, error) {
	descriptor, err := v1authoritypostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}

func CutoverApplyContract() (platformcutover.ApplyContract, error) {
	descriptor, err := adoptionpostgres.NewCutoverContract()
	if err != nil {
		return platformcutover.ApplyContract{}, err
	}
	return platformcutover.NewApplyContract(descriptor.ID(), descriptor.Query())
}

type V1AdoptionCompiler = adoptionapp.Compiler
type V1AdoptionMapping = adoptionapp.Mapping
type V1AdoptionPlan = adoptionapp.Plan
type V1AdoptionSnapshot = adoptionapp.Snapshot

var EncodeV1AdoptionPlan = adoptionapp.EncodePlan

func NewV1AdoptionCompiler() (*V1AdoptionCompiler, error) {
	return adoptionapp.NewCompiler(normalizeV1Email, normalizeV1Portal)
}

func normalizeV1Email(raw string) (string, error) {
	value, err := identifier.NormalizeEmailV1(raw)
	return value.Normalized(), err
}

func normalizeV1Portal(raw string) (string, error) {
	value, err := identifier.NormalizePortalV1(raw)
	return value.Normalized(), err
}

type ProfileCommands = profileapp.Commands
type ProfileReferencePort = profileapp.ReferencePort
type ProfileReferenceEvidence = profileapp.ReferenceEvidence
type ProfileRetirementTarget = profileapp.RetirementTarget
type ProfileResourceKind = profileapp.ResourceKind

const (
	ProfileResourceUser         = profileapp.ResourceUser
	ProfileResourceOrganization = profileapp.ResourceOrganization
	ProfileResourceRole         = profileapp.ResourceRole
)

func NewProfileCommands(
	references ProfileReferencePort,
) (*ProfileCommands, error) {
	commands, err := profilepostgres.NewCommandStore()
	if err != nil {
		return nil, err
	}
	return profileapp.NewCommands(commands, references)
}
