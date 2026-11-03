package iam

import (
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profile"
	adoptionapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/adoption"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/identifier"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

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
