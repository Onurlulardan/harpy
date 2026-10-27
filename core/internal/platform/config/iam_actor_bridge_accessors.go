package config

import (
	"crypto/ed25519"

	actorassertionv1 "github.com/Onurlulardan/harpy/contracts/actorassertion/v1"
)

type IAMActorBridgeTLSFiles struct {
	certificate AbsolutePath
	privateKey  AbsolutePath
	serverCA    AbsolutePath
}

func (config IAMActorBridge) Validate() error {
	if !config.valid || len(config.privateKey) != ed25519.PrivateKeySize ||
		len(config.publicKeys) < 1 || len(config.publicKeys) > 8 ||
		!exactActorBridgeDomains(config.targets) || config.clientCert.Value() == "" ||
		config.clientKey.Value() == "" || config.serverCA.Value() == "" {
		return errInvalidIAMAPIConfig
	}
	active, exists := config.publicKeys[config.keyID]
	if !exists || len(active) != ed25519.PublicKeySize ||
		!active.Equal(ed25519.PublicKey(config.privateKey[ed25519.SeedSize:])) {
		return errInvalidIAMAPIConfig
	}
	if _, err := actorassertionv1.NewVerifier(config.publicKeys); err != nil {
		return errInvalidIAMAPIConfig
	}
	return nil
}

func (config IAMActorBridge) SigningKey() (
	string,
	ed25519.PrivateKey,
	error,
) {
	if err := config.Validate(); err != nil {
		return "", nil, err
	}
	return config.keyID, append(ed25519.PrivateKey(nil), config.privateKey...), nil
}

// Close removes the startup-only signing key from every shallow config copy.
func (config IAMActorBridge) Close() {
	clear(config.privateKey)
}

func (config IAMActorBridge) PublicKeys() (
	map[string]ed25519.PublicKey,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	result := make(map[string]ed25519.PublicKey, len(config.publicKeys))
	for keyID, publicKey := range config.publicKeys {
		result[keyID] = append(ed25519.PublicKey(nil), publicKey...)
	}
	return result, nil
}

func (config IAMActorBridge) Targets() (
	map[actorassertionv1.Domain]IAMActorBridgeTarget,
	error,
) {
	if err := config.Validate(); err != nil {
		return nil, err
	}
	result := make(map[actorassertionv1.Domain]IAMActorBridgeTarget, len(config.targets))
	for _, target := range config.targets {
		result[target.domain] = target
	}
	return result, nil
}

func (target IAMActorBridgeTarget) Domain() actorassertionv1.Domain { return target.domain }
func (target IAMActorBridgeTarget) BaseURL() string                 { return target.baseURL }
func (target IAMActorBridgeTarget) ServerName() string              { return target.serverName }

func (config IAMActorBridge) ClientTLSFiles() (IAMActorBridgeTLSFiles, error) {
	if err := config.Validate(); err != nil {
		return IAMActorBridgeTLSFiles{}, err
	}
	return IAMActorBridgeTLSFiles{certificate: config.clientCert,
		privateKey: config.clientKey, serverCA: config.serverCA}, nil
}

func (files IAMActorBridgeTLSFiles) CertificateFile() string { return files.certificate.Value() }
func (files IAMActorBridgeTLSFiles) PrivateKeyFile() string  { return files.privateKey.Value() }
func (files IAMActorBridgeTLSFiles) ServerCAFile() string    { return files.serverCA.Value() }
