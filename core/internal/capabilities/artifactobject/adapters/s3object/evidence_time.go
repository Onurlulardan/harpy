package s3object

import (
	"errors"
	"time"

	artifactobject "github.com/Onurlulardan/harpy/core/internal/capabilities/artifactobject/api"
)

func (store *Store) evidenceNow() (time.Time, error) {
	value := store.clock.Now()
	if value.IsZero() || value.Location() != time.UTC || value.Year() < 2000 {
		return time.Time{}, errors.Join(artifactobject.ErrOutcomeUnknown,
			ErrProviderEvidence)
	}
	return value, nil
}
