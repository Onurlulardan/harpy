package postgres

import (
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/notifications/ports"
)

func scanProvider(row scanner) (ports.ProviderConfig, error) {
	var value ports.ProviderConfig
	var id, version, secret string
	var metadata, digest []byte
	err := row.Scan(&id, &version, &secret, &value.Kind, &value.Name,
		&value.Enabled, &value.Revision, &metadata, &digest)
	if err != nil {
		return ports.ProviderConfig{}, err
	}
	value.ID, err = decodeID(id)
	if err == nil {
		value.VersionID, err = decodeID(version)
	}
	if err == nil {
		value.SecretID, err = decodeID(secret)
	}
	if err == nil {
		value.MetadataDigest, err = decodeDigest(digest)
	}
	if err == nil {
		decoder := json.NewDecoder(bytes.NewReader(metadata))
		decoder.DisallowUnknownFields()
		err = decoder.Decode(&value.Metadata)
	}
	if err != nil || value.Revision < 1 ||
		(value.Kind != ports.ProviderSMTP && value.Kind != ports.ProviderTelegram) {
		return ports.ProviderConfig{}, fmt.Errorf("decode notification provider")
	}
	return value, nil
}

func encodeMetadata(value ports.ProviderMetadata) ([]byte, error) {
	return json.Marshal(value)
}
