package operator

import "time"

func runLocalArtifacts() error {
	return prepare(
		"/run/release/current/bootstrap-authority.json",
		"/run/installation", "/run/dev-authority",
		time.Now().UTC().Truncate(time.Second),
	)
}
