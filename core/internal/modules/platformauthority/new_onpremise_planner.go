package platformauthority

import (
	"crypto/ed25519"
	"fmt"

	recoverysigningv1 "github.com/Onurlulardan/harpy/contracts/recoverysigning/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/localkey"
	postgresadapter "github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/api"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/application"
)

func NewOnPremisePlanner(
	dependencies OnPremisePlannerDependencies,
) (api.ProtectedAuthorityPlanner, error) {
	if dependencies.Clock == nil || len(dependencies.RecoveryControlSeed) != ed25519.SeedSize {
		return nil, fmt.Errorf("on-premise authority planner input is invalid")
	}
	bundle, err := recoverysigningv1.Parse(dependencies.RecoveryTrustBundle)
	if err != nil {
		return nil, fmt.Errorf("parse recovery authority trust bundle: %w", err)
	}
	kid, found := bundle.KeyID(recoverysigningv1.PurposeRecoveryControl)
	public, publicFound := bundle.PublicKey(recoverysigningv1.PurposeRecoveryControl)
	private := ed25519.NewKeyFromSeed(dependencies.RecoveryControlSeed)
	defer clear(private)
	if !found || !publicFound ||
		!private.Public().(ed25519.PublicKey).Equal(ed25519.PublicKey(public[:])) {
		return nil, fmt.Errorf("recovery authority signer does not match trust bundle")
	}
	signer, err := localkey.New(kid, private)
	if err != nil {
		return nil, err
	}
	store, err := postgresadapter.NewStore(dependencies.Pool, dependencies.Clock)
	if err != nil {
		return nil, err
	}
	return application.NewProtectedPlanner(store, signer, dependencies.Clock)
}
