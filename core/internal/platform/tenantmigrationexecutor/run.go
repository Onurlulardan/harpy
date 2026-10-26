package tenantmigrationexecutor

import (
	"context"
	"time"

	"github.com/google/uuid"
)

func (executor *Executor) Run(ctx context.Context, gate Admission) error {
	poll := time.NewTicker(IdlePollInterval)
	readyHeartbeat := time.NewTicker(ReadyHeartbeat)
	defer poll.Stop()
	defer readyHeartbeat.Stop()
	var profileID uuid.UUID
	defer func() { executor.retire(context.WithoutCancel(ctx), profileID) }()
	for {
		if gate.Draining() {
			return nil
		}
		permit, admitted := gate.Admit()
		if !admitted {
			executor.retire(context.WithoutCancel(ctx), profileID)
			profileID = uuid.Nil
			select {
			case <-ctx.Done():
				return nil
			case <-poll.C:
			}
			continue
		}
		if profileID == uuid.Nil {
			var err error
			profileID, err = executor.register(ctx)
			if err != nil {
				return err
			}
		}
		processed, err := executor.processOne(ctx, permit, profileID)
		if err != nil {
			return err
		}
		if processed {
			continue
		}
		select {
		case <-ctx.Done():
			return nil
		case <-readyHeartbeat.C:
			if _, err := executor.store.HeartbeatReady(ctx, profileID, executor.readyTTL); err != nil {
				return err
			}
		case <-poll.C:
		}
	}
}

func (executor *Executor) register(ctx context.Context) (uuid.UUID, error) {
	profileID, err := executor.newID()
	if err != nil {
		return uuid.Nil, err
	}
	_, err = executor.store.RegisterReady(ctx, profileID, executor.release, executor.readyTTL)
	return profileID, err
}

func (executor *Executor) retire(ctx context.Context, profileID uuid.UUID) {
	if profileID != uuid.Nil {
		_ = executor.store.RetireReady(ctx, profileID)
	}
}
