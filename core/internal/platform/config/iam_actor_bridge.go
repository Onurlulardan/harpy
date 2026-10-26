package config

import (
	"crypto/ed25519"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"io"
)

const (
	iamActorBridgePrefix            = "HARPY_IAM_V1_ACTOR_BRIDGE_"
	iamActorBridgeKeyIDKey          = iamActorBridgePrefix + "SIGNING_KEY_ID"
	iamActorBridgeSeedFileKey       = iamActorBridgePrefix + "SIGNING_SEED_FILE"
	iamActorBridgePublicKeysFileKey = iamActorBridgePrefix + "PUBLIC_KEYS_FILE"
	iamActorBridgeTargetsFileKey    = iamActorBridgePrefix + "TARGETS_FILE"
	iamActorBridgeClientCertKey     = iamActorBridgePrefix + "CLIENT_CERT_FILE"
	iamActorBridgeClientKeyKey      = iamActorBridgePrefix + "CLIENT_KEY_FILE"
	iamActorBridgeServerCAKey       = iamActorBridgePrefix + "SERVER_CA_FILE"
)

type IAMActorBridge struct {
	keyID      string
	privateKey ed25519.PrivateKey
	publicKeys map[string]ed25519.PublicKey
	targets    map[string]IAMActorBridgeTarget
	clientCert AbsolutePath
	clientKey  AbsolutePath
	serverCA   AbsolutePath
	valid      bool
}

type actorBridgePublicKeyDocument struct {
	Version int                    `json:"version"`
	Keys    []actorBridgePublicKey `json:"keys"`
}

type actorBridgePublicKey struct {
	KeyID     string `json:"key_id"`
	PublicKey string `json:"public_key"`
}

func loadIAMActorBridge(values environmentValues) (IAMActorBridge, error) {
	keyID, found := values.lookup(iamActorBridgeKeyIDKey)
	if !found || keyID == "" {
		return IAMActorBridge{}, configError(iamActorBridgeKeyIDKey, "is required")
	}
	privateKey, err := loadActorBridgePrivateKey(values, keyID)
	if err != nil {
		return IAMActorBridge{}, err
	}
	publicKeys, err := loadActorBridgePublicKeys(values)
	if err != nil {
		clear(privateKey)
		return IAMActorBridge{}, err
	}
	activePublic, exists := publicKeys[keyID]
	if !exists || subtle.ConstantTimeCompare(
		activePublic, privateKey[ed25519.SeedSize:],
	) != 1 {
		clear(privateKey)
		return IAMActorBridge{}, configError(
			iamActorBridgePublicKeysFileKey, "does not contain the active signing key",
		)
	}
	targets, err := loadActorBridgeTargets(values)
	if err != nil {
		clear(privateKey)
		return IAMActorBridge{}, err
	}
	clientCert, err := loadAbsolutePath(values, iamActorBridgeClientCertKey)
	if err != nil {
		clear(privateKey)
		return IAMActorBridge{}, err
	}
	clientKey, err := loadAbsolutePath(values, iamActorBridgeClientKeyKey)
	if err != nil {
		clear(privateKey)
		return IAMActorBridge{}, err
	}
	serverCA, err := loadAbsolutePath(values, iamActorBridgeServerCAKey)
	if err != nil {
		clear(privateKey)
		return IAMActorBridge{}, err
	}
	return IAMActorBridge{keyID: keyID, privateKey: privateKey,
		publicKeys: publicKeys, targets: targets, clientCert: clientCert,
		clientKey: clientKey, serverCA: serverCA, valid: true}, nil
}

func loadActorBridgePrivateKey(
	values environmentValues,
	keyID string,
) (ed25519.PrivateKey, error) {
	path, found := values.lookup(iamActorBridgeSeedFileKey)
	if !found {
		return nil, configError(iamActorBridgeSeedFileKey, "is required")
	}
	encoded, err := readSecretFile(iamActorBridgeSeedFileKey, path, 128)
	if err != nil {
		return nil, err
	}
	seed, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil || len(seed) != ed25519.SeedSize ||
		base64.RawURLEncoding.EncodeToString(seed) != encoded {
		clear(seed)
		return nil, configError(iamActorBridgeSeedFileKey, "must contain one canonical Ed25519 seed")
	}
	privateKey := ed25519.NewKeyFromSeed(seed)
	clear(seed)
	if _, err := actorBridgeVerifier(map[string]ed25519.PublicKey{
		keyID: ed25519.PublicKey(privateKey[ed25519.SeedSize:]),
	}); err != nil {
		clear(privateKey)
		return nil, configError(iamActorBridgeKeyIDKey, "is invalid")
	}
	return privateKey, nil
}

func loadActorBridgePublicKeys(
	values environmentValues,
) (map[string]ed25519.PublicKey, error) {
	path, found := values.lookup(iamActorBridgePublicKeysFileKey)
	if !found {
		return nil, configError(iamActorBridgePublicKeysFileKey, "is required")
	}
	document, err := readSecretFile(iamActorBridgePublicKeysFileKey, path, secretFileLimit)
	if err != nil {
		return nil, err
	}
	if rejectDuplicateJSONMembers([]byte(document)) != nil {
		return nil, configError(iamActorBridgePublicKeysFileKey, "contains ambiguous JSON")
	}
	decoder := json.NewDecoder(newStringReader(document))
	decoder.DisallowUnknownFields()
	var raw actorBridgePublicKeyDocument
	if decoder.Decode(&raw) != nil || decoder.Decode(&struct{}{}) != io.EOF ||
		raw.Version != 1 || len(raw.Keys) < 1 || len(raw.Keys) > 8 {
		return nil, configError(iamActorBridgePublicKeysFileKey, "must satisfy iam-v1-actor-public-keys.v1")
	}
	keys := make(map[string]ed25519.PublicKey, len(raw.Keys))
	for _, entry := range raw.Keys {
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(entry.PublicKey)
		if decodeErr != nil || len(decoded) != ed25519.PublicKeySize ||
			base64.RawURLEncoding.EncodeToString(decoded) != entry.PublicKey {
			return nil, configError(iamActorBridgePublicKeysFileKey, "contains an invalid public key")
		}
		if _, duplicate := keys[entry.KeyID]; duplicate {
			return nil, configError(iamActorBridgePublicKeysFileKey, "contains a duplicate key ID")
		}
		keys[entry.KeyID] = ed25519.PublicKey(decoded)
	}
	if _, err := actorBridgeVerifier(keys); err != nil {
		return nil, configError(iamActorBridgePublicKeysFileKey, "contains an invalid key set")
	}
	return keys, nil
}
