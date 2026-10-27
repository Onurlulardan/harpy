package systemtransfer

import (
	tabularmapping "github.com/Onurlulardan/harpy/core/internal/capabilities/tabularmapping/api"
	publicresource "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/systemresource"
)

func BuildStoredMappingProgram(resource publicresource.Resource,
	document []byte) (tabularmapping.Program, error) {
	if !resource.Valid() {
		return tabularmapping.Program{}, ErrInvalidMapping
	}
	program, err := tabularmapping.Compile(document)
	if err != nil || program.Contract() != tabularmapping.ContractV2 {
		return tabularmapping.Program{}, ErrInvalidMapping
	}
	seen := make(map[string]struct{}, len(program.Targets()))
	for _, target := range program.Targets() {
		if _, parseErr := publicresource.ParseID(target); parseErr != nil ||
			protectedCoreTarget(resource, target) {
			return tabularmapping.Program{}, ErrForbiddenTarget
		}
		seen[target] = struct{}{}
	}
	for _, required := range createRequiredTargets(resource) {
		if _, found := seen[required]; !found {
			return tabularmapping.Program{}, ErrInviteRequired
		}
	}
	return program, nil
}
