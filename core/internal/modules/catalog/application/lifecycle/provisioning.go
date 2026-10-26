package lifecycle

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"time"

	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/componentowner"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
	catalogdomain "github.com/Onurlulardan/harpy/core/internal/modules/catalog/domain/lifecycle"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func (application *Application) provisionOwnerSteps(
	ctx context.Context,
	scope platformtx.TenantScope,
	operation catalogdomain.Operation,
) (time.Time, bool, error) {
	if application == nil || application.clock == nil {
		return time.Time{}, false, ErrUnavailable
	}
	sealed, err := application.loadPayloadEvidence(ctx, scope, operation.Plan)
	if err != nil {
		return time.Time{}, false, err
	}
	payloads := make(map[string]catalogpackagev1.OwnerPayload, len(sealed.OwnerPayloads))
	for _, payload := range sealed.OwnerPayloads {
		payloads[payload.ComponentID] = payload
	}
	started := application.clock.Now().UTC().Truncate(time.Microsecond)
	var waitUntil time.Time
	for index, pin := range operation.Plan.Components {
		payload, found := payloads[pin.ComponentID.String()]
		if !found || payload.Owner != pin.Owner ||
			payload.Contract != pin.Contract ||
			payload.ContentSHA256 != pin.ContentSHA256 {
			return time.Time{}, false, ErrUnavailable
		}
		owner, resolveErr := application.owners.Resolve(pin.Owner)
		provisioner, available := owner.(componentowner.Provisioner)
		if resolveErr != nil || !available {
			return time.Time{}, false, ErrOwnerNotV2
		}
		component := componentowner.Component{
			ID: pin.ComponentID.String(), Owner: pin.Owner,
			Generation: pin.OwnerGeneration, Contract: pin.Contract,
			ContentSHA256: pin.ContentSHA256,
		}
		request := componentowner.ProvisionRequest{
			PackageID:      operation.Plan.PackageID.String(),
			PackageSlug:    operation.Plan.Slug,
			PackageVersion: operation.Plan.Version.String(),
			ReleaseID:      sealed.ReleaseID, Component: component,
			Document: append([]byte(nil), payload.Document...),
			IdempotencyKey: fmt.Sprintf(
				"%s:%s:provision",
				operation.ID.String(), pin.ComponentID.String(),
			),
		}
		receipt, provisionErr := provisioner.Provision(ctx, scope, request)
		if err := classifyProvisionResult(provisionErr, receipt, request); err != nil {
			return time.Time{}, false, err
		}
		delete(payloads, pin.ComponentID.String())
		now := application.clock.Now().UTC().Truncate(time.Microsecond)
		wait, stop := provisionTickDecision(
			receipt.Ready, receipt.RetryAt, started, now,
			index+1 < len(operation.Plan.Components),
		)
		waitUntil = earliestProvisionWait(waitUntil, wait)
		if stop {
			if waitUntil.IsZero() {
				return now, false, nil
			}
			return floorProvisionRetry(waitUntil, now), false, nil
		}
	}
	if len(payloads) != 0 {
		return time.Time{}, false, ErrUnavailable
	}
	now := application.clock.Now().UTC().Truncate(time.Microsecond)
	return floorProvisionRetry(waitUntil, now), waitUntil.IsZero(), nil
}

func (application *Application) loadPayloadEvidence(
	ctx context.Context,
	scope platformtx.TenantScope,
	plan catalogdomain.Plan,
) (packageregistry.SealedVersion, error) {
	reader, err := application.registry.Bind(ctx, scope)
	if err != nil || reader == nil {
		return packageregistry.SealedVersion{}, ErrUnavailable
	}
	payloadReader, ok := reader.(packageregistry.PayloadReader)
	if !ok {
		return packageregistry.SealedVersion{}, ErrUnavailable
	}
	sealed, err := payloadReader.VersionPayloads(
		ctx, plan.PackageID.String(), plan.Version.String(),
	)
	if err != nil {
		return packageregistry.SealedVersion{}, ErrUnavailable
	}
	manifest, err := application.verifySealed(sealed)
	digest := sha256.Sum256(manifest.Canonical())
	if err != nil || manifest.PackageID != plan.PackageID ||
		manifest.Version.Compare(plan.Version) != 0 ||
		hex.EncodeToString(digest[:]) != plan.PackageSHA256 {
		return packageregistry.SealedVersion{}, ErrUnavailable
	}
	return sealed.Clone(), nil
}

func classifyProvisionResult(
	err error,
	receipt componentowner.ProvisionReceipt,
	request componentowner.ProvisionRequest,
) error {
	if err != nil {
		return err
	}
	if !provisionReceiptMatches(receipt, request) {
		return ErrOwnerReceiptStale
	}
	return nil
}

func provisionReceiptMatches(
	receipt componentowner.ProvisionReceipt,
	request componentowner.ProvisionRequest,
) bool {
	if !validStableID(receipt.ID) ||
		receipt.PackageID != request.PackageID ||
		receipt.PackageSlug != request.PackageSlug ||
		receipt.PackageVersion != request.PackageVersion ||
		receipt.ReleaseID != request.ReleaseID ||
		receipt.Component != request.Component ||
		(receipt.State != componentowner.ProvisionAvailable &&
			receipt.State != componentowner.ProvisionStaged) ||
		!canonicalProvisionTime(receipt.CompletedAt) {
		return false
	}
	if receipt.Ready {
		return receipt.RetryAt.IsZero()
	}
	return canonicalProvisionTime(receipt.RetryAt) &&
		receipt.RetryAt.After(receipt.CompletedAt)
}

func canonicalProvisionTime(value time.Time) bool {
	return !value.IsZero() && value.Location() == time.UTC &&
		value.Nanosecond()%1000 == 0
}
