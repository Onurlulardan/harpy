package platformauthority

import (
	"context"
	"crypto/ed25519"
	"fmt"
	"os"
	"runtime"

	recoveryv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/localkey"
	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type DevelopmentBindingDependencies struct {
	Pool              *DeliveryPool
	Clock             DeliveryClock
	EnvironmentID     string
	RecoveryTrustPath string
	RecoverySeedPath  string
	AckTrustPath      string
	ReleaseTrustPath  string
	JournalPath       string
	AckSeedPath       string
}

func BindDevelopmentInstallation(
	ctx context.Context,
	dependencies DevelopmentBindingDependencies,
) (api.BindingReceipt, error) {
	if ctx == nil || ctx.Err() != nil || dependencies.Pool == nil ||
		dependencies.Clock == nil {
		return api.BindingReceipt{}, fmt.Errorf(
			"development authority binding dependencies are invalid",
		)
	}
	store, err := postgresadapter.NewStore(dependencies.Pool, dependencies.Clock)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	input, err := application.ReadDevelopmentBindingInput(
		ctx, store, dependencies.EnvironmentID,
	)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	signer, proof, err := developmentBindingProof(dependencies, input)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	defer clear(proof)
	registryDependencies := DeliveryDependencies{
		Clock: dependencies.Clock, Mode: DeliveryDevelopment,
		AckTrustPath:      dependencies.AckTrustPath,
		ReleaseTrustPath:  dependencies.ReleaseTrustPath,
		RecoveryTrustPath: dependencies.RecoveryTrustPath,
		LocalJournalPath:  dependencies.JournalPath,
		LocalAckSeedPath:  dependencies.AckSeedPath,
	}
	trust, err := verifyDeliveryTrust(registryDependencies, ports.RegistryDevelopment)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	registry, err := newDeliveryRegistry(ctx, registryDependencies, trust)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	service, err := application.NewBindingService(
		store, registry, signer, dependencies.Clock, ports.RegistryDevelopment,
	)
	if err != nil {
		return api.BindingReceipt{}, err
	}
	return service.BindOnPremise(ctx, api.OnPremiseBindingRequest{
		BindingID: input.RegistryBindingID, EnvironmentID: dependencies.EnvironmentID,
		FinalizationReceiptSHA256: input.FinalizationReceiptSHA256,
		RecoveryTrustBundleSHA256: input.RecoveryTrustBundleSHA256,
		ProofBytes:                proof,
	})
}

func developmentBindingProof(
	dependencies DevelopmentBindingDependencies,
	input ports.DevelopmentBindingInput,
) (ports.ProposalSigner, []byte, error) {
	bundleDocument, err := readDeliveryArtifact(dependencies.RecoveryTrustPath)
	if err != nil {
		return nil, nil, err
	}
	defer clear(bundleDocument)
	bundle, err := recoveryv1.Parse(bundleDocument)
	if err != nil || bundle.EnvironmentID() != dependencies.EnvironmentID ||
		bundle.SHA256() != input.RecoveryTrustBundleSHA256 {
		return nil, nil, fmt.Errorf("development recovery trust is invalid")
	}
	seed, err := readDevelopmentBindingSeed(dependencies.RecoverySeedPath)
	if err != nil {
		clear(seed)
		return nil, nil, fmt.Errorf("development recovery signer is invalid")
	}
	defer clear(seed)
	private := ed25519.NewKeyFromSeed(seed)
	defer clear(private)
	kid, found := bundle.KeyID(recoveryv1.PurposeRecoveryControl)
	public, publicFound := bundle.PublicKey(recoveryv1.PurposeRecoveryControl)
	if !found || !publicFound ||
		!private.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(public[:])) {
		return nil, nil, fmt.Errorf("development recovery signer is untrusted")
	}
	signer, err := localkey.New(kid, private)
	if err != nil {
		return nil, nil, err
	}
	proof, err := application.SignDevelopmentBindingProof(
		application.DevelopmentBindingProofInput{
			EnvironmentID:             dependencies.EnvironmentID,
			LicenseBindingID:          input.RegistryBindingID,
			FinalizationReceiptSHA256: input.FinalizationReceiptSHA256,
			RecoveryTrustBundleSHA256: input.RecoveryTrustBundleSHA256,
			ControlKID:                kid, OperationID: input.OperationID,
		}, private,
	)
	if err != nil {
		return nil, nil, err
	}
	return signer, proof, nil
}

func readDevelopmentBindingSeed(path string) ([]byte, error) {
	info, err := os.Lstat(path)
	if err != nil || !info.Mode().IsRegular() || info.Mode()&os.ModeSymlink != 0 ||
		info.Size() != ed25519.SeedSize ||
		runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, fmt.Errorf("development recovery signer file is invalid")
	}
	seed, err := os.ReadFile(path)
	if err != nil || len(seed) != ed25519.SeedSize {
		clear(seed)
		return nil, fmt.Errorf("read development recovery signer")
	}
	return seed, nil
}
