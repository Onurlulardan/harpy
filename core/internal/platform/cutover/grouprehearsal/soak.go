package grouprehearsal

import "time"

type SoakEvidence struct {
	Cohort           Cohort
	StartedAt        time.Time
	EndedAt          time.Time
	EligibleRequests uint64
	ComparedRequests uint64
	MismatchCount    uint64
	Metrics          Metrics
}

func EvaluateSoak(value SoakEvidence) error {
	minimum, err := MinimumSoak(value.Cohort)
	if err != nil || value.StartedAt.IsZero() || !value.EndedAt.After(value.StartedAt) ||
		value.EndedAt.Sub(value.StartedAt) < minimum || value.EligibleRequests == 0 ||
		value.ComparedRequests != value.EligibleRequests || value.MismatchCount != 0 {
		return ErrInvalid
	}
	fence, err := value.Metrics.RequiresFence()
	if err != nil {
		return err
	}
	if fence {
		return ErrFence
	}
	return nil
}
