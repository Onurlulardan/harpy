package usage

import "time"

type ReserveExecution struct {
	Request ReserveRequest
	Now     time.Time
}

type ReservePlanner func(ReservationSnapshot, *ReservationReplay) (ReservePlan, error)

type ReservePlan struct {
	Result      ReserveResult
	Decision    *DecisionRecord
	Reservation *Reservation
	Event       *ThresholdEventIntent
}

type AdjustExecution struct {
	Request AdjustGaugeRequest
	Now     time.Time
}

type AdjustPlanner func(GaugeSnapshot, *GaugeReplay) (AdjustPlan, error)

type AdjustPlan struct {
	Result   AdjustGaugeResult
	Decision *DecisionRecord
	Entry    *GaugeEntry
	Event    *ThresholdEventIntent
}

type StartExecution struct {
	Request StartEffectRequest
	Now     time.Time
}

type TransitionExecution struct {
	Request TransitionRequest
	Now     time.Time
}

type ExpireExecution struct {
	Request ExpireRequest
	Now     time.Time
}

type ReservationPlanner func(Reservation, bool) (ReservationPlan, error)

type ReservationPlan struct {
	Result     ReservationResult
	Effect     *EffectStart
	Transition *ReservationTransition
}
