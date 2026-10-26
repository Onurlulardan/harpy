package installationdelivery

import "context"

func DeleteSecret(
	ctx context.Context, store Store, command StateCommand,
) (Transition, error) {
	if unusable(ctx, store) || !command.valid() {
		return Transition{}, ErrInvalid
	}
	result, err := store.DeleteSecret(ctx, command)
	if err != nil {
		return Transition{}, err
	}
	if !result.valid(StateSecretDeleted) {
		return Transition{}, ErrUnavailable
	}
	return result, nil
}
