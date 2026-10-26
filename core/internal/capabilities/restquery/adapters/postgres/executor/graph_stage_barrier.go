package executor

import (
	"context"
	"sort"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/planner"
)

type graphStageBarrier struct {
	policy   planner.Policy
	expected []int
	requests chan graphStageSubmission
	stages   []chan struct{}
	done     chan struct{}
	once     sync.Once
	mutex    sync.RWMutex
	failure  error
}

func newGraphStageBarrier(
	policy planner.Policy,
	expected []int,
) (*graphStageBarrier, error) {
	if len(expected) < 1 || len(expected) > 5 || validatePolicy(policy) != nil {
		return nil, ErrContract
	}
	for _, count := range expected {
		if count < 1 || count > 4 {
			return nil, ErrContract
		}
	}
	stages := make([]chan struct{}, len(expected))
	for index := range stages {
		stages[index] = make(chan struct{})
	}
	return &graphStageBarrier{
		policy: policy, expected: append([]int(nil), expected...),
		requests: make(chan graphStageSubmission, 16), stages: stages,
		done: make(chan struct{}),
	}, nil
}

func (barrier *graphStageBarrier) run(
	ctx context.Context,
	programs *graphProgramRun,
) error {
	if barrier == nil || ctx == nil || programs == nil || programs.done == nil {
		return ErrContract
	}
	tracker, err := planner.NewTracker(barrier.policy)
	if err != nil {
		return barrier.finish(ErrContract)
	}
	pending := make([][]graphStageSubmission, len(barrier.expected))
	bound := graphStatementBound{}
	for stage, count := range barrier.expected {
		for len(pending[stage]) < count {
			select {
			case submission := <-barrier.requests:
				if barrier.accept(pending, stage, submission) != nil {
					return barrier.failPending(pending, ErrContract)
				}
			case <-programs.done:
				runErr := programs.err
				if runErr == nil {
					runErr = ErrContract
				}
				return barrier.failPending(pending, runErr)
			case <-ctx.Done():
				return barrier.failPending(pending, ctx.Err())
			}
		}
		works := pending[stage]
		sort.Slice(works, func(left, right int) bool {
			return works[left].work.ordinal < works[right].work.ordinal
		})
		if duplicateWorkOrdinal(works) {
			return barrier.failPending(pending, ErrContract)
		}
		bound, err = bound.add(works)
		if err != nil {
			return barrier.failPending(pending, err)
		}
		tracker, err = runGraphStage(ctx, barrier.policy, tracker, works)
		if err == nil {
			close(barrier.stages[stage])
		}
		for _, submission := range works {
			submission.reply <- err
		}
		pending[stage] = nil
		if err != nil {
			return barrier.failPending(pending, err)
		}
	}
	return barrier.finish(nil)
}

func (barrier *graphStageBarrier) accept(
	pending [][]graphStageSubmission,
	current int,
	submission graphStageSubmission,
) error {
	stage := submission.work.stage
	if submission.validate() != nil ||
		stage < current || stage >= len(pending) {
		return ErrContract
	}
	pending[stage] = append(pending[stage], submission)
	if len(pending[stage]) > barrier.expected[stage] {
		return ErrContract
	}
	return nil
}

func duplicateWorkOrdinal(values []graphStageSubmission) bool {
	for index := 1; index < len(values); index++ {
		if values[index-1].work.ordinal == values[index].work.ordinal {
			return true
		}
	}
	return false
}
