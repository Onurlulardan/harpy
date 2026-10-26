package iam

import (
	"fmt"
	"time"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
	platformactorbridge "github.com/Onurlulardan/harpy/core/internal/platform/actorbridge"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func NewActorBridgeRuntime(
	settings config.IAMAPI,
	coordinator *platformtx.Coordinator,
) (*ActorBridgeRuntime, error) {
	bridgeSettings, err := settings.ActorBridge()
	if err != nil {
		return nil, fmt.Errorf("load IAM V1 actor bridge settings: %w", err)
	}
	defer bridgeSettings.Close()
	keyID, privateKey, err := bridgeSettings.SigningKey()
	if err != nil {
		return nil, fmt.Errorf("load IAM V1 actor bridge signer: %w", err)
	}
	tlsFiles, err := bridgeSettings.ClientTLSFiles()
	if err != nil {
		clear(privateKey)
		return nil, fmt.Errorf("load IAM V1 actor bridge TLS settings: %w", err)
	}
	targets, err := actorBridgeTargets(bridgeSettings)
	if err != nil {
		clear(privateKey)
		return nil, err
	}
	signer, err := actorassertionv1.NewProductionSigner(keyID, privateKey)
	clear(privateKey)
	if err != nil {
		return nil, fmt.Errorf("build IAM V1 actor bridge signer: %w", err)
	}
	ownedSigner := true
	defer func() {
		if ownedSigner {
			signer.Close()
		}
	}()
	authorizer, err := NewActorBridgeIssuerIAM()
	if err != nil {
		return nil, err
	}
	authority, err := newActorBridgeAuthority()
	if err != nil {
		return nil, fmt.Errorf("build IAM V1 actor bridge authority: %w", err)
	}
	issuer, err := platformactorbridge.NewIssuer(
		coordinator, authorizer, authority, signer, time.Now,
	)
	if err != nil {
		return nil, fmt.Errorf("build IAM V1 actor bridge issuer: %w", err)
	}
	identity, err := platformactorbridge.LoadClientIdentity(
		platformactorbridge.ClientTLSFiles{CertificateFile: tlsFiles.CertificateFile(),
			PrivateKeyFile: tlsFiles.PrivateKeyFile(), ServerCAFile: tlsFiles.ServerCAFile()},
		time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("load IAM V1 actor bridge mTLS identity: %w", err)
	}
	client, err := platformactorbridge.NewClient(identity, targets, time.Now)
	if err != nil {
		return nil, fmt.Errorf("build IAM V1 actor bridge client: %w", err)
	}
	forwarder, err := platformactorbridge.NewForwarder(issuer, client)
	if err != nil {
		client.Close()
		issuer.Close()
		return nil, fmt.Errorf("build IAM V1 actor bridge forwarder: %w", err)
	}
	runtime := &ActorBridgeRuntime{forwarder: forwarder}
	if !runtime.Ready() {
		runtime.Close()
		return nil, fmt.Errorf("IAM V1 actor bridge is not ready")
	}
	ownedSigner = false
	return runtime, nil
}

func actorBridgeTargets(
	settings config.IAMActorBridge,
) (map[actorassertionv1.Domain]platformactorbridge.Target, error) {
	configured, err := settings.Targets()
	if err != nil {
		return nil, fmt.Errorf("load IAM V1 actor bridge targets: %w", err)
	}
	targets := make(map[actorassertionv1.Domain]platformactorbridge.Target, len(configured))
	for domain, entry := range configured {
		target, targetErr := platformactorbridge.NewTarget(
			domain, entry.BaseURL(), entry.ServerName(),
		)
		if targetErr != nil {
			return nil, fmt.Errorf("build IAM V1 actor bridge target: %w", targetErr)
		}
		targets[domain] = target
	}
	return targets, nil
}
