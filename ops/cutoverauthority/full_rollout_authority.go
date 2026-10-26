package main

import (
	"fmt"
	"time"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
	"github.com/Onurlulardan/harpy/internal/attestation"
)

func loadHistoricalAuthorityBundle(
	paths authorityArtifactPaths, now time.Time,
) (verifiedAuthorityBundle, error) {
	document, err := readArtifact(paths.AuthorityOperation, maximumArtifactBytes)
	if err != nil {
		return verifiedAuthorityBundle{}, err
	}
	var operation cutoverv1.AuthorityAdoptionOperation
	if err := attestation.DecodeStrict(document, &operation); err != nil {
		return verifiedAuthorityBundle{}, fmt.Errorf("authority operation verification failed")
	}
	issuedAt, err := time.Parse(time.RFC3339Nano, operation.IssuedAt)
	if err != nil || issuedAt.Location() != time.UTC || issuedAt.After(now) {
		return verifiedAuthorityBundle{}, fmt.Errorf("authority operation time is invalid")
	}
	return loadAuthorityBundle(paths, issuedAt)
}
