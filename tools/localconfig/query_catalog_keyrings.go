package main

import (
	"encoding/base64"
	"time"
)

func addQueryKeyring(set *artifactSet, _ time.Time) error {
	currentID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	previousID, err := randomLowerID("local-")
	if err != nil {
		return err
	}
	currentKey, err := randomBase64(32)
	if err != nil {
		return err
	}
	previousKey, err := randomBase64(32)
	if err != nil {
		return err
	}
	if err := addJSON(set, "rest-query-cursor-keyring.json", map[string]any{
		"version":  1,
		"current":  map[string]string{"kid": currentID, "key": currentKey},
		"previous": map[string]string{"kid": previousID, "key": previousKey},
	}); err != nil {
		return err
	}
	return nil
}
