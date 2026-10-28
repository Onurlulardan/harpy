package main

import (
	"fmt"
	"os"
	"path/filepath"
	"time"
)

type ensureResult struct {
	Generated bool
	Output    string
	Manifest  setManifest
}

func ensureArtifactSet(output string, now time.Time) (ensureResult, error) {
	resolved, err := resolveOutputPath(output)
	if err != nil {
		return ensureResult{}, err
	}
	if err := recoverArtifactUpgrade(resolved, now); err != nil {
		return ensureResult{}, err
	}
	manifest, current, err := inspectArtifactSet(resolved, now, false)
	if err != nil {
		return ensureResult{}, err
	}
	if current {
		return ensureResult{Output: resolved, Manifest: manifest}, nil
	}
	if _, statErr := os.Lstat(resolved); statErr == nil {
		if result, upgraded, upgradeErr := upgradeRecoveryArtifactSet(
			resolved, now,
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeAuthorityArtifactSet(
			resolved, now,
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeExecutorEnvironmentArtifactSet(
			resolved, now,
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeRecoveryPathArtifactSet(
			resolved, now,
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeBrowserOriginArtifactSet(
			resolved, now, localGenerationOptions(),
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeAIMCPRuntimeArtifactSet(
			resolved, now, localGenerationOptions(),
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		if result, upgraded, upgradeErr := upgradeMigratorCatalogArtifactSet(
			resolved, now, localGenerationOptions(),
		); upgradeErr != nil {
			return ensureResult{}, upgradeErr
		} else if upgraded {
			return result, nil
		}
		_, intact, inspectErr := inspectArtifactSet(resolved, now, true)
		if inspectErr != nil {
			return ensureResult{}, inspectErr
		}
		if intact {
			return ensureResult{}, fmt.Errorf(
				"local config expired; destroy persistent volumes, then run with -reset",
			)
		}
		return ensureResult{}, fmt.Errorf(
			"local config does not match the current catalog or failed integrity validation; " +
				"automatic replacement is refused; explicit reset verifies legacy integrity",
		)
	} else if !os.IsNotExist(statErr) {
		return ensureResult{}, fmt.Errorf("inspect existing artifact set: %w", statErr)
	}
	set, err := generateArtifactSet(now)
	if err != nil {
		return ensureResult{}, fmt.Errorf("generate artifact set: %w", err)
	}
	manifest, err = buildSetManifest(set, now)
	if err != nil {
		return ensureResult{}, err
	}
	if err := publishArtifactSet(resolved, set, manifest); err != nil {
		return ensureResult{}, err
	}
	return ensureResult{
		Generated: true, Output: resolved, Manifest: manifest,
	}, nil
}

func resolveOutputPath(output string) (string, error) {
	resolved, err := filepath.Abs(filepath.Clean(output))
	if err != nil {
		return "", fmt.Errorf("resolve output: %w", err)
	}
	root := filepath.VolumeName(resolved) + string(os.PathSeparator)
	if resolved == root || filepath.Base(resolved) == "." {
		return "", fmt.Errorf("output must be a dedicated non-root directory")
	}
	if info, statErr := os.Lstat(resolved); statErr == nil {
		if info.Mode()&os.ModeSymlink != 0 || !info.IsDir() {
			return "", fmt.Errorf("output must be a real directory")
		}
	} else if !os.IsNotExist(statErr) {
		return "", fmt.Errorf("inspect output: %w", statErr)
	}
	if err := os.MkdirAll(filepath.Dir(resolved), 0o700); err != nil {
		return "", fmt.Errorf("create output parent: %w", err)
	}
	return resolved, nil
}
