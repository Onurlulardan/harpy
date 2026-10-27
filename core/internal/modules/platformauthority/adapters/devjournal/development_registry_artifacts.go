package devjournal

import (
	"crypto/ed25519"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
)

const (
	DevelopmentAckTrustFilename    = "protected-authority-ack-trust-bundle.v1.json"
	DevelopmentReleaseRootFilename = "protected-authority-release-trust-root.json"
	developmentAckKID              = "local-authority-ack-1"
)

func ReconcileDevelopmentRegistryArtifacts(
	root string,
	ackSeedPath string,
	releaseSeedPath string,
	now time.Time,
) error {
	current := filepath.Join(root, "current")
	if !validDevelopmentDirectory(current) || now.IsZero() {
		return fmt.Errorf("development authority registry directory is invalid")
	}
	ackPrivate, err := readDevelopmentSeed(ackSeedPath)
	if err != nil {
		return err
	}
	defer clear(ackPrivate)
	releasePrivate, err := readDevelopmentSeed(releaseSeedPath)
	if err != nil {
		return err
	}
	defer clear(releasePrivate)
	if ackPrivate.Public().(ed25519.PublicKey).Equal(
		releasePrivate.Public().(ed25519.PublicKey),
	) {
		return fmt.Errorf("development authority keys are not purpose-separated")
	}
	rootPath := filepath.Join(current, DevelopmentReleaseRootFilename)
	trustPath := filepath.Join(current, DevelopmentAckTrustFilename)
	if _, err = os.Lstat(rootPath); err == nil {
		if _, trustErr := os.Lstat(trustPath); trustErr == nil {
			return verifyDevelopmentRegistryArtifacts(
				rootPath, trustPath, ackPrivate, releasePrivate, now,
			)
		} else if !os.IsNotExist(trustErr) {
			return trustErr
		}
		matched, matchErr := developmentReleaseRootMatches(
			rootPath, releasePrivate.Public().(ed25519.PublicKey), now,
		)
		if matchErr != nil || !matched {
			return fmt.Errorf("development authority release trust is invalid")
		}
		notBefore := now.UTC().Truncate(time.Second).Add(-time.Minute)
		trustDocument, buildErr := buildDevelopmentAckTrust(
			releasePrivate, ackPrivate.Public().(ed25519.PublicKey),
			notBefore, notBefore.AddDate(10, 0, 0),
		)
		if buildErr != nil {
			return buildErr
		}
		if writeErr := writeDevelopmentExclusive(
			trustPath, trustDocument, 0o400,
		); writeErr != nil {
			return writeErr
		}
		return verifyDevelopmentRegistryArtifacts(
			rootPath, trustPath, ackPrivate, releasePrivate, now,
		)
	} else if !os.IsNotExist(err) {
		return err
	}
	if _, trustErr := os.Lstat(trustPath); !os.IsNotExist(trustErr) {
		return fmt.Errorf("development authority artifact set is incomplete")
	}
	notBefore := now.UTC().Truncate(time.Second).Add(-time.Minute)
	signUntil := notBefore.AddDate(10, 0, 0)
	rootDocument, err := buildDevelopmentReleaseRoot(
		releasePrivate.Public().(ed25519.PublicKey), notBefore, signUntil,
	)
	if err != nil {
		return err
	}
	trustDocument, err := buildDevelopmentAckTrust(
		releasePrivate, ackPrivate.Public().(ed25519.PublicKey),
		notBefore, signUntil,
	)
	if err != nil {
		return err
	}
	if err = writeDevelopmentExclusive(rootPath, rootDocument, 0o400); err != nil {
		return err
	}
	if err = writeDevelopmentExclusive(trustPath, trustDocument, 0o400); err != nil {
		return err
	}
	return verifyDevelopmentRegistryArtifacts(
		rootPath, trustPath, ackPrivate, releasePrivate, now,
	)
}

func verifyDevelopmentRegistryArtifacts(
	rootPath string,
	trustPath string,
	ackPrivate ed25519.PrivateKey,
	releasePrivate ed25519.PrivateKey,
	now time.Time,
) error {
	releasePublic := releasePrivate.Public().(ed25519.PublicKey)
	matched, err := developmentReleaseRootMatches(rootPath, releasePublic, now)
	if err != nil {
		return err
	}
	trustDocument, err := readDevelopmentRegular(trustPath, artifact.MaximumBytes)
	if err != nil || !matched {
		return fmt.Errorf("development authority release trust is invalid")
	}
	trust, err := artifact.VerifyTrustBundle(
		trustDocument, artifact.RegistryDevelopment, now, releasePublic,
	)
	if err != nil {
		return fmt.Errorf("verify development authority trust bundle: %w", err)
	}
	kid, active, found := trust.ActiveKey()
	if !found || kid != developmentAckKID ||
		!active.Equal(ackPrivate.Public().(ed25519.PublicKey)) {
		return fmt.Errorf("development authority acknowledgement signer mismatches")
	}
	return nil
}

func readDevelopmentSeed(path string) (ed25519.PrivateKey, error) {
	seed, err := readDevelopmentRegular(path, ed25519.SeedSize)
	if err != nil || len(seed) != ed25519.SeedSize {
		return nil, fmt.Errorf("development authority seed is invalid")
	}
	defer clear(seed)
	return ed25519.NewKeyFromSeed(seed), nil
}

func validDevelopmentDirectory(path string) bool {
	info, err := os.Lstat(path)
	return err == nil && info.IsDir() && info.Mode()&os.ModeSymlink == 0
}
