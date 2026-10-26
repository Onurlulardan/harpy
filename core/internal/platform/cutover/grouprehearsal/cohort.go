package grouprehearsal

import (
	"fmt"
	"time"
)

type Cohort string

const (
	CohortInternal   Cohort = "internal"
	CohortOne        Cohort = "1_percent"
	CohortFive       Cohort = "5_percent"
	CohortTwentyFive Cohort = "25_percent"
	CohortFifty      Cohort = "50_percent"
	CohortAll        Cohort = "100_percent"
)

var cohortOrder = []Cohort{
	CohortInternal, CohortOne, CohortFive, CohortTwentyFive, CohortFifty, CohortAll,
}

func MinimumSoak(cohort Cohort) (time.Duration, error) {
	switch cohort {
	case CohortInternal:
		return 24 * time.Hour, nil
	case CohortOne, CohortFive:
		return 48 * time.Hour, nil
	case CohortTwentyFive, CohortFifty:
		return 72 * time.Hour, nil
	case CohortAll:
		return 7 * 24 * time.Hour, nil
	default:
		return 0, fmt.Errorf("%w: unknown cohort", ErrInvalid)
	}
}

func CohortOrder() []Cohort { return append([]Cohort(nil), cohortOrder...) }
