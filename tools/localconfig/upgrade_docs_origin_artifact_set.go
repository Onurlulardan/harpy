package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

const docsOriginUpgradeSourceRevision = 14

var previousLocalBrowserOrigins = []byte(
	"HARPY_IAM_ALLOWED_ORIGINS=" +
		"[\"http://localhost:5173\",\"http://127.0.0.1:5173\"]\n",
)

func upgradeDocsOriginArtifactSet(
	output string,
	now time.Time,
	options generationOptions,
) (ensureResult, bool, error) {
	legacy, intact, err := inspectManifestBackedSet(output, func(value setManifest) error {
		return validateBrowserOriginUpgradeManifest(
			value, now, options, docsOriginUpgradeSourceRevision,
		)
	})
	if err != nil || !intact {
		return ensureResult{}, false, err
	}
	set := newArtifactSetFor(options)
	for _, file := range legacy.Files {
		content, readErr := os.ReadFile(filepath.Join(output, file.Name))
		if readErr != nil {
			return ensureResult{}, false, readErr
		}
		set.files[file.Name] = content
	}
	set.expiresAt, err = parseCanonicalTimestamp(legacy.ExpiresAt)
	if err != nil {
		return ensureResult{}, false, err
	}
	if options.profile == profileLocal {
		apiEnvironment := set.files["api.env"]
		if bytes.Count(apiEnvironment, previousLocalBrowserOrigins) != 1 ||
			bytes.Contains(apiEnvironment, currentLocalBrowserOrigins) {
			return ensureResult{}, false, fmt.Errorf(
				"local API environment is not eligible for docs origin upgrade",
			)
		}
		set.files["api.env"] = bytes.Replace(
			apiEnvironment, previousLocalBrowserOrigins, currentLocalBrowserOrigins, 1,
		)
	}
	manifest, err := buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, false, err
	}
	if err = publishBrowserOriginUpgrade(output, set, manifest); err != nil {
		return ensureResult{}, false, err
	}
	return ensureResult{Generated: true, Output: output, Manifest: manifest}, true, nil
}
