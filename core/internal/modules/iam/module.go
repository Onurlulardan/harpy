package iam

import (
	profilepostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profile"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
)

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
