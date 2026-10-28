package main

import (
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
)

func validateCurrentManifestCatalog(
	files []manifestFile,
	profile generationProfile,
	licenseKind string,
) error {
	if err := validateManifestFileCatalog(files); err != nil {
		return err
	}
	expected := newArtifactSetFor(generationOptions{
		profile: profile, licenseKind: licenseKind,
	}).expectedArtifactNames()
	if len(files) != len(expected) {
		return fmt.Errorf("manifest does not match the current artifact catalog")
	}
	for index, file := range files {
		if file.Name != expected[index] {
			return fmt.Errorf("manifest does not match the current artifact catalog")
		}
	}
	return nil
}

func validateManifestFileCatalog(files []manifestFile) error {
	if len(files) == 0 {
		return fmt.Errorf("manifest file catalog is empty")
	}
	previous := ""
	for _, file := range files {
		if !canonicalArtifactBaseName(file.Name) ||
			(previous != "" && file.Name <= previous) ||
			file.Size < 1 || !canonicalSHA256(file.SHA256) {
			return fmt.Errorf("manifest file catalog is invalid")
		}
		previous = file.Name
	}
	return nil
}

func canonicalArtifactBaseName(name string) bool {
	return name != "" && name != "." && name != ".." && name != manifestFileName &&
		name == filepath.Base(name) && !strings.ContainsAny(name, `/\\`)
}

func canonicalSHA256(value string) bool {
	if len(value) != 64 {
		return false
	}
	digest, err := hex.DecodeString(value)
	return err == nil && len(digest) == 32 && hex.EncodeToString(digest) == value
}
