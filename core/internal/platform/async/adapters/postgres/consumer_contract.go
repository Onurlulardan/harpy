package postgres

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	effectsv1 "github.com/Onurlulardan/harpy/contracts/effects/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
)

func snapshotEnvelope(
	value eventsv1.Envelope,
	producers eventsv1.ProducerCatalog,
) (eventsv1.Envelope, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return eventsv1.Envelope{}, err
	}
	return eventsv1.DecodeEnvelope(document, producers)
}

func snapshotEventReceipt(
	value eventsv1.EffectReceipt,
	envelope eventsv1.Envelope,
	producers eventsv1.ProducerCatalog,
	consumers eventsv1.ConsumerCatalog,
) (eventsv1.EffectReceipt, string, error) {
	if err := value.Validate(envelope, producers, consumers); err != nil {
		return eventsv1.EffectReceipt{}, "", err
	}
	document, err := json.Marshal(value)
	if err != nil {
		return eventsv1.EffectReceipt{}, "", err
	}
	snapshot, err := eventsv1.DecodeEffectReceipt(document, envelope, producers, consumers)
	if err != nil {
		return eventsv1.EffectReceipt{}, "", err
	}
	digest, err := canonicalDigest(document)
	return snapshot, digest, err
}

func snapshotBootstrapReceipt(
	value eventsv1.BootstrapReceipt,
	consumers eventsv1.ConsumerCatalog,
) (eventsv1.BootstrapReceipt, string, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return eventsv1.BootstrapReceipt{}, "", err
	}
	snapshot, err := eventsv1.DecodeBootstrapReceipt(document, consumers)
	if err != nil {
		return eventsv1.BootstrapReceipt{}, "", err
	}
	digest, err := canonicalDigest(document)
	return snapshot, digest, err
}

func snapshotExternalReceipt(value effectsv1.Receipt) (effectsv1.Receipt, error) {
	document, err := json.Marshal(value)
	if err != nil {
		return effectsv1.Receipt{}, err
	}
	return effectsv1.DecodeReceipt(document)
}

func canonicalDigest(document []byte) (string, error) {
	canonical, err := canonicaljsonv1.Canonicalize(document, eventsv1.MaximumDocumentBytes)
	if err != nil {
		return "", fmt.Errorf("canonicalize async receipt: %w", err)
	}
	digest := sha256.Sum256(canonical)
	return hex.EncodeToString(digest[:]), nil
}
