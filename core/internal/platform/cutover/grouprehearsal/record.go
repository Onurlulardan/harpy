package grouprehearsal

import (
	"context"
	"fmt"
)

func Record(
	ctx context.Context, recorder Recorder, value PersistedRehearsal,
) (Acceptance, error) {
	if ctx == nil || ctx.Err() != nil || recorder == nil ||
		validatePersisted(value) != nil {
		return Acceptance{}, ErrInvalid
	}
	digest := digestPersisted(value)
	persisted, err := recorder.Record(ctx, value, digest)
	if err != nil {
		return Acceptance{}, fmt.Errorf("record billing Forms group rehearsal: %w", err)
	}
	if persisted != digest {
		return Acceptance{}, ErrInvalid
	}
	return Acceptance{digest: digest}, nil
}
