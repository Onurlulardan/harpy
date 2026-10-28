package main

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type generationProfile string

const (
	profileLocal      generationProfile = "local"
	profileForgeGuard generationProfile = "forgeguard"
)

type generationOptions struct {
	profile        generationProfile
	publicURL      string
	forgeGuardURL  string
	licenseKind    string
	inputDirectory string
	inputDigest    string
}

func localGenerationOptions() generationOptions {
	return generationOptions{profile: profileLocal}
}

func forgeGuardGenerationOptions(
	publicURL string,
	forgeGuardURL string,
	licenseKind string,
	inputDirectory string,
) (generationOptions, error) {
	options := generationOptions{
		profile: profileForgeGuard, publicURL: strings.TrimSpace(publicURL),
		forgeGuardURL:  strings.TrimRight(strings.TrimSpace(forgeGuardURL), "/"),
		licenseKind:    strings.TrimSpace(licenseKind),
		inputDirectory: filepath.Clean(strings.TrimSpace(inputDirectory)),
	}
	if !canonicalHTTPSURL(options.publicURL) ||
		!canonicalHTTPSURL(options.forgeGuardURL) ||
		(options.licenseKind != "connected" && options.licenseKind != "airgap") ||
		!filepath.IsAbs(options.inputDirectory) {
		return generationOptions{}, fmt.Errorf("forgeguard generation options are invalid")
	}
	digest, err := forgeGuardInputDigest(options.inputDirectory, options.licenseKind)
	if err != nil {
		return generationOptions{}, err
	}
	options.inputDigest = digest
	return options, nil
}

func canonicalHTTPSURL(value string) bool {
	parsed, err := url.Parse(value)
	return err == nil && parsed.Scheme == "https" && parsed.Host != "" &&
		parsed.User == nil && parsed.RawQuery == "" && parsed.Fragment == "" &&
		parsed.String() == value
}

func (options generationOptions) digest() string {
	if options.profile == profileLocal {
		return ""
	}
	value := strings.Join([]string{
		string(options.profile), options.publicURL, options.forgeGuardURL,
		options.licenseKind, options.inputDigest,
	}, "\x00")
	digest := sha256.Sum256([]byte(value))
	return hex.EncodeToString(digest[:])
}

func forgeGuardInputDigest(directory, licenseKind string) (string, error) {
	names := forgeGuardInputNames(licenseKind)
	sort.Strings(names)
	hasher := sha256.New()
	for _, name := range names {
		path := filepath.Join(directory, name)
		info, err := os.Lstat(path)
		if err != nil || !info.Mode().IsRegular() ||
			info.Mode()&os.ModeSymlink != 0 || info.Size() < 1 || info.Size() > 1<<20 {
			return "", fmt.Errorf("forgeguard input %s is invalid", name)
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return "", fmt.Errorf("read forgeguard input %s: %w", name, err)
		}
		_, _ = hasher.Write([]byte(name))
		_, _ = hasher.Write([]byte{0})
		_, _ = hasher.Write(content)
		_, _ = hasher.Write([]byte{0})
	}
	return hex.EncodeToString(hasher.Sum(nil)), nil
}
