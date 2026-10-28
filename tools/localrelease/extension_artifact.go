package main

import "fmt"

func exactExtensionArtifact() (artifact, error) {
	value, err := newArtifact("extension-manifest.json", extensionManifest())
	if err != nil {
		return artifact{}, err
	}
	if value.digest != extensionManifestHash {
		return artifact{}, fmt.Errorf(
			"extension manifest digest differs: got %s want %s",
			value.digest, extensionManifestHash,
		)
	}
	return value, nil
}
