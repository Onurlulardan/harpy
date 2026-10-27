package catalogprojection

import (
	"bytes"
	"regexp"
	"strings"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/google/uuid"
)

var (
	semverPattern = regexp.MustCompile(
		`^(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)` +
			`(?:-[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?` +
			`(?:\+[0-9A-Za-z-]+(?:\.[0-9A-Za-z-]+)*)?$`,
	)
	idempotencyPattern = regexp.MustCompile(`^[A-Za-z0-9_.:-]{1,128}$`)
	digestPattern      = regexp.MustCompile(`^[0-9a-f]{64}$`)
	identityPattern    = regexp.MustCompile(`^[a-z][a-z0-9-]{1,127}$`)
)

type verifiedRequest struct {
	request         contract.ProvisionRequest
	packageID       uuid.UUID
	releaseID       uuid.UUID
	materialization Materialization
}

func (application *Application) verify(
	request contract.ProvisionRequest,
) (verifiedRequest, error) {
	packageID, packageErr := canonicalUUID(request.PackageID, false)
	releaseID, releaseErr := canonicalUUID(request.ReleaseID, true)
	if packageErr != nil || releaseErr != nil ||
		!validPackageSlug(request.PackageSlug) ||
		!semverPattern.MatchString(request.PackageVersion) ||
		!idempotencyPattern.MatchString(request.IdempotencyKey) ||
		!application.validComponent(request.Component, true) {
		return verifiedRequest{}, ErrInvalid
	}
	materialization, err := application.decoder.Decode(
		request.Document, request.Component,
	)
	if err != nil || !validMaterialization(materialization, request.Document) ||
		materialization.ModuleSlug != request.PackageSlug {
		return verifiedRequest{}, ErrInvalid
	}
	return verifiedRequest{
		request: request.Clone(), packageID: packageID, releaseID: releaseID,
		materialization: materialization.Clone(),
	}, nil
}

func canonicalUUID(value string, requireV7 bool) (uuid.UUID, error) {
	parsed, err := uuid.Parse(value)
	if err != nil || parsed == uuid.Nil || parsed.String() != strings.ToLower(value) ||
		requireV7 && parsed.Version() != 7 {
		return uuid.Nil, ErrInvalid
	}
	return parsed, nil
}

func (application *Application) validComponent(
	component contract.Component,
	requireGeneration bool,
) bool {
	_, err := canonicalUUID(component.ID, true)
	if err != nil || component.Owner != application.owner ||
		component.Contract != application.contract ||
		!digestPattern.MatchString(component.ContentSHA256) {
		return false
	}
	if requireGeneration {
		return component.Generation > 0
	}
	return component.Generation >= 0
}

func validMaterialization(value Materialization, document []byte) bool {
	return value.Kind != "" && len(value.Kind) <= 32 &&
		validPackageSlug(value.ModuleSlug) &&
		identityPattern.MatchString(value.Identity) &&
		bytes.Equal(value.Document, document) &&
		value.SemanticSHA256 != [32]byte{} &&
		value.ItemCount > 0 && value.AuxiliaryCount >= 0
}

func validPackageSlug(value string) bool {
	contract, err := catalogpackagev1.FirstPartySourceContract(value)
	return err == nil && catalogpackagev1.ValidFirstPartySourceContract(contract)
}
