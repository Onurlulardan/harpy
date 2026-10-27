package jobs

import (
	"time"

	"github.com/Onurlulardan/harpy/core/internal/modules/work/domain"
)

func mustAdapterID(value string) domain.ID {
	id, err := domain.ParseID(value)
	if err != nil {
		panic(err)
	}
	return id
}

func mustAdapterTime() time.Time {
	return time.Date(2026, 7, 19, 10, 0, 0, 0, time.UTC)
}
