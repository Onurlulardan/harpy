package main

import "time"

func addGeneralKeyrings(set *artifactSet, now time.Time) error {
	steps := []func(*artifactSet, time.Time) error{
		addQueryAndCatalogKeyrings,
		addArtifactKEKKeyrings,
		addNotificationsKeyrings,
	}
	for _, step := range steps {
		if err := step(set, now); err != nil {
			return err
		}
	}
	return nil
}
