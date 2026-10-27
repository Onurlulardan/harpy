package executor

import "context"

func (barrier *graphStageBarrier) submit(
	ctx context.Context,
	work graphStageWork,
) error {
	if barrier == nil || ctx == nil || work.validate() != nil {
		return ErrContract
	}
	return barrier.send(ctx, graphStageSubmission{
		work: work, reply: make(chan error, 1),
	})
}

func (barrier *graphStageBarrier) wait(
	ctx context.Context,
	stage int,
	ordinal int,
) error {
	if barrier == nil || ctx == nil {
		return ErrContract
	}
	submission := graphStageSubmission{
		work:  graphStageWork{stage: stage, ordinal: ordinal},
		reply: make(chan error, 1), waiting: true,
	}
	if submission.validate() != nil {
		return ErrContract
	}
	return barrier.send(ctx, submission)
}

func (barrier *graphStageBarrier) send(
	ctx context.Context,
	submission graphStageSubmission,
) error {
	select {
	case barrier.requests <- submission:
	case <-barrier.done:
		return barrier.result()
	case <-ctx.Done():
		return ctx.Err()
	}
	select {
	case err := <-submission.reply:
		return err
	case <-barrier.done:
		return barrier.result()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (barrier *graphStageBarrier) await(
	ctx context.Context,
	stage int,
) error {
	if barrier == nil || ctx == nil || stage < 0 || stage >= len(barrier.stages) {
		return ErrContract
	}
	select {
	case <-barrier.stages[stage]:
		return nil
	case <-barrier.done:
		return barrier.result()
	case <-ctx.Done():
		return ctx.Err()
	}
}
