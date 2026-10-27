package api

import (
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

const notificationConsumerDomain = "harpy:notifications:legacy-consumer:v1\x00"

type notificationConsumerManifest struct {
	Version     int    `json:"version"`
	Consumer    string `json:"consumer"`
	Route       string `json:"route"`
	Replacement string `json:"replacement"`
	ExpiresAt   string `json:"expires_at"`
}

type notificationConsumerVerifier struct {
	keyring config.NotificationsConsumerKeyring
	now     func() time.Time
}

func (verifier notificationConsumerVerifier) Verify(manifestValues,
	signatureValues []string,
) ([32]byte, error) {
	manifestWire, valid := exactNotificationHeader(manifestValues, 4096)
	signatureWire, signatureValid := exactNotificationHeader(signatureValues, 512)
	if !valid || !signatureValid || verifier.now == nil {
		return [32]byte{}, fmt.Errorf("notification consumer proof is invalid")
	}
	raw, err := base64.RawURLEncoding.Strict().DecodeString(manifestWire)
	if err != nil || len(raw) == 0 || len(raw) > 2048 ||
		base64.RawURLEncoding.EncodeToString(raw) != manifestWire {
		return [32]byte{}, fmt.Errorf("notification consumer manifest is invalid")
	}
	manifest, err := parseNotificationConsumerManifest(raw, verifier.now().UTC())
	if err != nil {
		return [32]byte{}, err
	}
	digest := sha256.Sum256(raw)
	keyID, encodedSignature, found := strings.Cut(signatureWire, ".")
	signature, decodeErr := base64.RawURLEncoding.Strict().DecodeString(encodedSignature)
	if !found || strings.Contains(encodedSignature, ".") || decodeErr != nil ||
		len(signature) != ed25519.SignatureSize ||
		base64.RawURLEncoding.EncodeToString(signature) != encodedSignature {
		return [32]byte{}, fmt.Errorf("notification consumer signature is invalid")
	}
	key, err := verifier.keyring.Resolve(keyID, manifest.Consumer, digest)
	if err != nil {
		return [32]byte{}, fmt.Errorf("notification consumer key is unavailable")
	}
	message := append([]byte(notificationConsumerDomain), raw...)
	verified := ed25519.Verify(key, message, signature)
	clear(message)
	if !verified {
		return [32]byte{}, fmt.Errorf("notification consumer signature is invalid")
	}
	return digest, nil
}

func parseNotificationConsumerManifest(raw []byte, now time.Time) (
	notificationConsumerManifest, error,
) {
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	var manifest notificationConsumerManifest
	if decoder.Decode(&manifest) != nil || decoder.Decode(&struct{}{}) != io.EOF {
		return manifest, fmt.Errorf("notification consumer manifest is invalid")
	}
	canonical, err := json.Marshal(manifest)
	expires, expiryErr := time.Parse(time.RFC3339, manifest.ExpiresAt)
	if err != nil || !bytes.Equal(canonical, raw) || manifest.Version != 1 ||
		manifest.Consumer == "" || manifest.Route != "/ws/notifications/{user_id}" ||
		manifest.Replacement != "/ws/notifications" || expiryErr != nil ||
		expires.Location() != time.UTC || expires.Format(time.RFC3339) != manifest.ExpiresAt ||
		!now.Before(expires) {
		return manifest, fmt.Errorf("notification consumer manifest is invalid")
	}
	return manifest, nil
}

func exactNotificationHeader(values []string, maximum int) (string, bool) {
	if len(values) != 1 {
		return "", false
	}
	return values[0], values[0] != "" &&
		len(values[0]) <= maximum && strings.TrimSpace(values[0]) == values[0] &&
		!strings.ContainsAny(values[0], "\r\n\t ")
}
