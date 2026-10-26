package installationdelivery

import "context"

func Complete(ctx context.Context, store Store, command StateCommand) (Transition, error) {
	if unusable(ctx, store) || !command.valid() {
		return Transition{}, ErrInvalid
	}
	result, err := store.Complete(ctx, command)
	if err != nil {
		return Transition{}, err
	}
	if !result.valid(StateCompleted) {
		return Transition{}, ErrUnavailable
	}
	return result, nil
}
