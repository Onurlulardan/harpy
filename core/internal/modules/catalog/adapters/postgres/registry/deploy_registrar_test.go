package registry

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	catalogpackagev1 "github.com/Onurlulardan/harpy/contracts/catalogpackage/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/catalog/api/packageregistry"
)

const deployPackageID = "018f7000-0000-7000-8000-000000000001"
const deployComponentID = "018f7000-0000-7000-8000-000000000002"
const deployReleaseID = "018f7000-0000-7000-8000-000000000003"

type deployVersionRow struct {
	value               packageregistry.SealedVersion
	payloadReleaseID    string
	payloadRegisteredAt time.Time
	err                 error
}

func (source deployVersionRow) Scan(destinations ...any) error {
	if source.err != nil {
		return source.err
	}
	if len(destinations) != 10 {
		return fmt.Errorf("deploy version destinations = %d", len(destinations))
	}
	*(destinations[0].(*string)) = source.value.PackageID
	*(destinations[1].(*string)) = source.value.Slug
	*(destinations[2].(*string)) = source.value.Version
	*(destinations[3].(*[]byte)) = append([]byte(nil), source.value.Manifest...)
	*(destinations[4].(*string)) = source.value.ManifestSHA256
	*(destinations[5].(*[]byte)) = append([]byte(nil), source.value.Signature...)
	*(destinations[6].(*string)) = source.value.KeyID
	*(destinations[7].(*string)) = source.value.ReleaseID
	*(destinations[8].(*time.Time)) = source.value.RegisteredAt
	releaseID := source.payloadReleaseID
	if releaseID == "" {
		releaseID = source.value.ReleaseID
	}
	registeredAt := source.payloadRegisteredAt
	if registeredAt.IsZero() {
		registeredAt = source.value.RegisteredAt
	}
	stored := make([]storedOwnerPayload, len(source.value.OwnerPayloads))
	for index := range source.value.OwnerPayloads {
		stored[index] = storedOwnerPayload{
			OwnerPayload: source.value.OwnerPayloads[index],
			ReleaseID:    releaseID,
			RegisteredAt: registeredAt,
		}
	}
	document, err := json.Marshal(stored)
	if err != nil {
		return err
	}
	payloads, err := canonicaljsonv1.Canonicalize(
		document, catalogpackagev1.MaximumPackageSetBytes,
	)
	if err != nil {
		return err
	}
	*(destinations[9].(*[]byte)) = payloads
	return nil
}

func TestDeployRegistrarRejectsInvalidEvidenceBeforeQuery(t *testing.T) {
	value := deployVersion(t)
	value.ManifestSHA256 = strings.Repeat("f", 64)
	called := false
	registrar := newDeployRegistrar(
		func(context.Context, string, ...any) DeployRow {
			called = true
			return deployVersionRow{}
		},
		func(error) bool { return false },
	)
	err := registrar.Register(context.Background(), value)
	if !errors.Is(err, packageregistry.ErrInvalid) || called {
		t.Fatalf("error = %v, query called = %t", err, called)
	}
}

func TestNewDeployRegistrarRequiresDatabasePort(t *testing.T) {
	if _, err := NewDeployRegistrar(nil, nil); !errors.Is(err, packageregistry.ErrUnavailable) {
		t.Fatalf("constructor error = %v", err)
	}
}
