package iam

import (
	"fmt"

	queryapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/restquery/systemresource"
	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

func identityDirectoryFromConfig(
	settings config.IAMAPI,
) (identityDirectoryKey, error) {
	keyring, err := settings.LoginDirectoryKeyring()
	if err != nil {
		return identityDirectoryKey{}, fmt.Errorf("load IAM directory keyring: %w", err)
	}
	keys, err := keyring.Keys()
	if err != nil || len(keys) == 0 {
		return identityDirectoryKey{}, fmt.Errorf("load IAM active directory key")
	}
	id, idErr := keys[0].ID()
	material, materialErr := keys[0].Material()
	active, activeErr := keys[0].Active()
	for index := range keys {
		keys[index].Destroy()
	}
	if idErr != nil || materialErr != nil || activeErr != nil || !active {
		clear(material[:])
		return identityDirectoryKey{}, fmt.Errorf("load IAM active directory key")
	}
	return identityDirectoryKey{id: id, material: material}, nil
}

func identityQueryKeysFromConfig(
	keyring config.CursorKeyring,
) (queryapp.CursorKeys, error) {
	current, err := keyring.Current()
	if err != nil {
		return queryapp.CursorKeys{}, fmt.Errorf("load REST Query current cursor key: %w", err)
	}
	previous, err := keyring.Previous()
	if err != nil {
		return queryapp.CursorKeys{}, fmt.Errorf("load REST Query previous cursor key: %w", err)
	}
	currentID, currentIDErr := current.KID()
	currentKey, currentKeyErr := current.AES256Key()
	previousID, previousIDErr := previous.KID()
	previousKey, previousKeyErr := previous.AES256Key()
	if currentIDErr != nil || currentKeyErr != nil ||
		previousIDErr != nil || previousKeyErr != nil {
		clear(currentKey[:])
		clear(previousKey[:])
		return queryapp.CursorKeys{}, fmt.Errorf("load REST Query cursor keys")
	}
	keys, err := queryapp.NewCursorKeys(
		currentID, currentKey, previousID, previousKey,
	)
	clear(currentKey[:])
	clear(previousKey[:])
	return keys, err
}
