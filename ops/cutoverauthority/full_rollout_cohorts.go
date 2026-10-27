package main

import (
	corecutover "github.com/Onurlulardan/harpy/core/public/cutover"
)

func decodeFullRolloutCohorts(values []fullRolloutCohort) (
	[]corecutover.FullRolloutSoakEvidence, error,
) {
	result := make([]corecutover.FullRolloutSoakEvidence, len(values))
	for index, value := range values {
		startedAt, err := parseFullRolloutTime(value.StartedAt)
		if err != nil {
			return nil, err
		}
		endedAt, err := parseFullRolloutTime(value.EndedAt)
		if err != nil {
			return nil, err
		}
		result[index] = corecutover.FullRolloutSoakEvidence{
			Cohort: corecutover.FullRolloutCohort(value.Cohort), StartedAt: startedAt,
			EndedAt: endedAt, EligibleRequests: value.EligibleRequests,
			ComparedRequests: value.ComparedRequests, MismatchCount: value.MismatchCount,
			Metrics: decodeFullRolloutMetrics(value.Metrics),
		}
	}
	return result, nil
}
