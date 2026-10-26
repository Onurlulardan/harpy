package operator

import (
	"context"
	"fmt"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/modules/tenancy"
	"github.com/Onurlulardan/harpy/core/internal/orchestration/environmentoperator"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	installationmanifest "github.com/Onurlulardan/harpy/core/internal/platform/installationmanifest"
)

type activationRuntime interface {
	FinalizeActivation(
		context.Context,
		tenancy.OperatorFinalizationRequest,
		tenancy.OperatorActivationCallback,
	) error
}

type InitialCredentialLoader func(context.Context) ([]byte, error)

type EnvironmentOperatorFinalizer struct {
	runtime activationRuntime
	keyring config.IAMDirectoryKeyring
	load    InitialCredentialLoader
}

func NewEnvironmentOperatorFinalizer(
	runtime activationRuntime,
	keyring config.IAMDirectoryKeyring,
	load InitialCredentialLoader,
) (*EnvironmentOperatorFinalizer, error) {
	if runtime == nil || load == nil {
		return nil, ErrEnvironmentFinalizerInvalid
	}
	keys, err := keyring.Keys()
	if err != nil {
		return nil, ErrEnvironmentFinalizerInvalid
	}
	for index := range keys {
		keys[index].Destroy()
	}
	return &EnvironmentOperatorFinalizer{runtime: runtime, keyring: keyring, load: load}, nil
}

func (finalizer *EnvironmentOperatorFinalizer) Finalize(
	ctx context.Context,
	request environmentoperator.FinalizationRequest,
) error {
	if finalizer == nil || ctx == nil || ctx.Err() != nil ||
		!request.Plan.IsValid() || request.ExpectedRevision < 1 ||
		request.AttemptID == [16]byte{} || request.LeaseTokenSHA256 == "" ||
		request.Plan.ManifestSHA256() != request.Manifest.SHA256() {
		return ErrEnvironmentFinalizerInvalid
	}
	input, err := finalizer.prepare(ctx, request.Manifest, request.Plan.CreatedAt())
	if err != nil {
		return err
	}
	callback, err := EnvironmentActivationFinalizer(input)
	if err != nil {
		return err
	}
	return finalizer.runtime.FinalizeActivation(ctx,
		tenancy.OperatorFinalizationRequest{Plan: request.Plan,
			LeaseTokenSHA256: request.LeaseTokenSHA256,
			ExpectedRevision: request.ExpectedRevision}, callback)
}

func (finalizer *EnvironmentOperatorFinalizer) prepare(
	ctx context.Context,
	manifest installationmanifest.Manifest,
	activatedAt time.Time,
) (EnvironmentFinalizerInput, error) {
	admin, tenant := manifest.FirstAdmin(), manifest.PrimaryTenant()
	directory, err := iam.PrepareInstallationDirectory(admin.Email, finalizer.keyring)
	if err != nil {
		return EnvironmentFinalizerInput{}, fmt.Errorf("prepare IAM directory evidence: %w", err)
	}
	presented, err := finalizer.load(ctx)
	if err != nil {
		return EnvironmentFinalizerInput{}, fmt.Errorf("load initial credential: %w", err)
	}
	defer clear(presented)
	credential, err := iam.PrepareInstallationCredential(presented)
	if err != nil {
		return EnvironmentFinalizerInput{}, fmt.Errorf("prepare initial credential: %w", err)
	}
	return EnvironmentFinalizerInput{TenantName: tenant.Name, TenantSlug: tenant.Slug,
		AdminEmail: admin.Email, AdminDisplayName: admin.DisplayName,
		DirectoryKeyID: directory.KeyID, IdentifierHMAC: directory.LookupHMAC,
		DirectorySourceSHA256: directory.SourceSHA256,
		InitialCredential:     credential, ActivatedAt: activatedAt}, nil
}

var _ environmentoperator.Finalizer = (*EnvironmentOperatorFinalizer)(nil)
