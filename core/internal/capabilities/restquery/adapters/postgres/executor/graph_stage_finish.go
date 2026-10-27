package executor

func (barrier *graphStageBarrier) failPending(
	pending [][]graphStageSubmission,
	err error,
) error {
	for _, stage := range pending {
		for _, submission := range stage {
			select {
			case submission.reply <- err:
			default:
			}
		}
	}
	return barrier.finish(err)
}

func (barrier *graphStageBarrier) finish(err error) error {
	if barrier == nil || barrier.done == nil {
		return ErrContract
	}
	barrier.once.Do(func() {
		barrier.mutex.Lock()
		barrier.failure = err
		barrier.mutex.Unlock()
		close(barrier.done)
	})
	return err
}

func (barrier *graphStageBarrier) result() error {
	if barrier == nil {
		return ErrContract
	}
	barrier.mutex.RLock()
	err := barrier.failure
	barrier.mutex.RUnlock()
	return err
}
