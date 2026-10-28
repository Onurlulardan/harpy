package main

import "time"

type artifactKEKTarget struct {
	file string
	kid  string
}

type artifactKEKDocument struct {
	KID string `json:"kid"`
	Key string `json:"key"`
}

type artifactKEKKeyringDocument struct {
	Version  int                   `json:"version"`
	Current  artifactKEKDocument   `json:"current"`
	Retained []artifactKEKDocument `json:"retained"`
}

var analyticsDashboardExportKEKTarget = artifactKEKTarget{
	"analytics-dashboard-export-fcart-kek-keyring.json", "local-analytics-export-",
}

var artifactKEKTargets = []artifactKEKTarget{
	{"documents-kek-keyring.json", "local-documents-"},
	analyticsDashboardExportKEKTarget,
	{"iam-system-export-fcart-kek-keyring.json", "local-iam-export-"},
	{"iam-system-import-source-fcart-kek-keyring.json", "local-iam-import-"},
}

func addArtifactKEKKeyrings(set *artifactSet, _ time.Time) error {
	for _, target := range artifactKEKTargets {
		if err := addArtifactKEKKeyring(set, target); err != nil {
			return err
		}
	}
	return nil
}

func addArtifactKEKKeyring(set *artifactSet, target artifactKEKTarget) error {
	kid, err := randomLowerID(target.kid)
	if err != nil {
		return err
	}
	key, err := randomBase64(32)
	if err != nil {
		return err
	}
	return addJSON(set, target.file, artifactKEKKeyringDocument{
		Version: 1, Current: artifactKEKDocument{KID: kid, Key: key},
		Retained: []artifactKEKDocument{},
	})
}
