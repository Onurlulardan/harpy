package config

import "os"

type PostRestoreReactivationObserver struct {
	runtime DedicatedRuntime
	input   ArtifactPath
}

func LoadPostRestoreReactivationObserver() (
	PostRestoreReactivationObserver,
	error,
) {
	return loadPostRestoreReactivationObserver(os.Environ())
}

func loadPostRestoreReactivationObserver(
	entries []string,
) (PostRestoreReactivationObserver, error) {
	values, runtimeConfig, err := loadDedicatedRuntime(
		entries, profilePostRestoreObserver,
		"HARPY_POST_RESTORE_REACTIVATION_POSTGRES_DSN_FILE",
		"harpy_reactivation_observer",
	)
	if err != nil {
		return PostRestoreReactivationObserver{}, err
	}
	input, err := loadReadArtifact(
		values, "HARPY_POST_RESTORE_REACTIVATION_INPUT_FILE", false,
	)
	if err != nil {
		return PostRestoreReactivationObserver{}, err
	}
	if err = validateDistinctArtifactPaths(
		"HARPY_POST_RESTORE_REACTIVATION_*",
		runtimeConfig.DSNFile(), input,
	); err != nil {
		return PostRestoreReactivationObserver{}, err
	}
	return PostRestoreReactivationObserver{
		runtime: runtimeConfig, input: input,
	}, nil
}

func (config PostRestoreReactivationObserver) Runtime() DedicatedRuntime {
	return config.runtime
}

func (config PostRestoreReactivationObserver) Input() ArtifactPath {
	return config.input
}
