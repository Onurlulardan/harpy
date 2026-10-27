package demomaterialization

import (
	"crypto/sha256"
	"regexp"
	"strings"

	contract "github.com/Onurlulardan/harpy/contracts/catalogcomponent/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/catalogpayload"
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
)

type verifiedRequest struct {
	request         contract.ProvisionRequest
	packageID       uuid.UUID
	releaseID       uuid.UUID
	materialization Materialization
}

func verifyRequest(
	request contract.ProvisionRequest,
) (verifiedRequest, error) {
	packageID, packageErr := canonicalUUID(request.PackageID, false)
	releaseID, releaseErr := canonicalUUID(request.ReleaseID, true)
	if packageErr != nil || releaseErr != nil ||
		!validPackageSlug(request.PackageSlug) ||
		!semverPattern.MatchString(request.PackageVersion) ||
		!idempotencyPattern.MatchString(request.IdempotencyKey) ||
		!validComponent(request.Component, true) {
		return verifiedRequest{}, ErrInvalid
	}
	verified, err := catalogpayload.Verify(
		request.Document, request.Component.ID, request.Component.ContentSHA256,
	)
	if err != nil {
		return verifiedRequest{}, ErrInvalid
	}
	payload := verified.Value()
	sourceSlug, sourceErr := catalogpackagev1.ParseFirstPartySourceContract(
		payload.SourceContract,
	)
	if sourceErr != nil || sourceSlug != request.PackageSlug ||
		payload.ModuleSlug != request.PackageSlug {
		return verifiedRequest{}, ErrInvalid
	}
	rowCount := 0
	for _, form := range payload.DemoForms {
		rowCount += len(form.Rows)
	}
	materialization := Materialization{
		ModuleSlug: payload.ModuleSlug, SourceContract: payload.SourceContract,
		Document: verified.CanonicalJSON(), SemanticSHA256: sha256.Sum256(request.Document),
		FormCount: len(payload.DemoForms), RowCount: rowCount,
	}
	if materialization.FormCount < 1 || materialization.RowCount < 1 {
		return verifiedRequest{}, ErrInvalid
	}
	return verifiedRequest{
		request: request.Clone(), packageID: packageID, releaseID: releaseID,
		materialization: materialization,
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

func validComponent(component contract.Component, requireGeneration bool) bool {
	_, err := canonicalUUID(component.ID, true)
	if err != nil || component.Owner != owner ||
		component.Contract != catalogContract() ||
		!digestPattern.MatchString(component.ContentSHA256) {
		return false
	}
	if requireGeneration {
		return component.Generation > 0
	}
	return component.Generation >= 0
}

func validPackageSlug(value string) bool {
	contractValue, err := catalogpackagev1.FirstPartySourceContract(value)
	return err == nil && catalogpackagev1.ValidFirstPartySourceContract(contractValue)
}
