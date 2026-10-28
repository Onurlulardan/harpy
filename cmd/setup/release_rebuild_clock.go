package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Onurlulardan/harpy/internal/attestation"
)

// preservedSourceClock keeps source-derived release stamps on rebuild.
// A fresh clock is used when no installed authority document matches the
// setup issue formula. Callers still mint a new trust root.
func preservedSourceClock(directory string, fresh releaseClock) (releaseClock, error) {
	installed, ok, err := installedDocumentClock(directory)
	if err != nil || !ok {
		return fresh, err
	}
	return installed, nil
}

func installedDocumentClock(directory string) (releaseClock, bool, error) {
	issued, expires, ok, err := installedDocumentStamps(directory, "authority-manifest.json")
	if err != nil || !ok {
		return releaseClock{}, false, err
	}
	clock := newReleaseClock(issued.Add(2 * time.Minute))
	if !clock.issued.Equal(issued) || !clock.releaseExpires.Equal(expires) {
		return releaseClock{}, false, nil
	}
	matched, err := cutoverMatchesClock(directory, clock)
	if err != nil || !matched {
		return releaseClock{}, false, err
	}
	return clock, true, nil
}

func cutoverMatchesClock(directory string, clock releaseClock) (bool, error) {
	issued, expires, ok, err := installedDocumentStamps(directory, "cutover-operation.json")
	if err != nil || !ok {
		return err == nil, err
	}
	return issued.Equal(clock.issued) && expires.Equal(clock.expires), nil
}

func installedDocumentStamps(directory, name string) (time.Time, time.Time, bool, error) {
	path := filepath.Join(directory, name)
	exists, err := installedFile(path)
	if err != nil || !exists {
		return time.Time{}, time.Time{}, false, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return time.Time{}, time.Time{}, false, err
	}
	var stamped struct {
		IssuedAt  string `json:"issued_at"`
		ExpiresAt string `json:"expires_at"`
	}
	if err = json.Unmarshal(body, &stamped); err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("read %s stamps: %w", name, err)
	}
	issued, err := time.Parse(time.RFC3339, stamped.IssuedAt)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("read %s issued_at: %w", name, err)
	}
	expires, err := time.Parse(time.RFC3339, stamped.ExpiresAt)
	if err != nil {
		return time.Time{}, time.Time{}, false, fmt.Errorf("read %s expires_at: %w", name, err)
	}
	return issued, expires, true, nil
}

func coverInstalledIssuedAt(keys releaseKeys, source releaseClock) (releaseKeys, error) {
	if !source.issued.Before(keys.clock.issued) {
		return keys, nil
	}
	var root trustedReleaseRoot
	if err := json.Unmarshal(keys.rootDocument, &root); err != nil {
		return releaseKeys{}, fmt.Errorf("read release trust root: %w", err)
	}
	if len(root.Keys) != 1 {
		return releaseKeys{}, fmt.Errorf("release trust root key count is invalid")
	}
	root.Keys[0].NotBefore = keys.clock.stamp(source.issued.Add(-time.Hour))
	document, err := attestation.MarshalCanonical(root)
	if err != nil {
		return releaseKeys{}, err
	}
	parsed, err := attestation.ParseTrustRoot(document)
	if err != nil {
		return releaseKeys{}, err
	}
	keys.root = parsed
	keys.rootDocument = document
	return keys, nil
}
