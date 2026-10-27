package executor

import "context"

func runGraphLifecycle(
	ctx context.Context,
	barrier *graphStageBarrier,
	programs *graphProgramRun,
) (error, error) {
	if ctx == nil || barrier == nil || programs == nil {
		return ErrContract, ErrContract
	}
	var barrierErr error
	var panicValue any
	func() {
		defer func() {
			if recovered := recover(); recovered != nil {
				panicValue = recovered
				_ = barrier.finish(ErrContract)
			}
		}()
		barrierErr = barrier.run(ctx, programs)
	}()
	programErr := programs.outcome()
	if panicValue != nil {
		panic(panicValue)
	}
	return barrierErr, programErr
}
