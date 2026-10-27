package application

import (
	"context"
	"errors"
	"fmt"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

var (
	ErrDeliveryAwaitingImport = errors.New("protected authority acknowledgement import is required")
	ErrDeliveryPermanent      = errors.New("protected authority delivery is permanently invalid")
)

type DeliveryProcessor struct {
	store        ports.Store
	registry     ports.Registry
	clock        ports.Clock
	trust        artifact.VerifiedTrustBundle
	registryKind artifact.RegistryKind
	finalizer    ports.OperationFinalizer
	airGap       bool
}

func NewDeliveryProcessor(
	store ports.Store,
	registry ports.Registry,
	clock ports.Clock,
	trust artifact.VerifiedTrustBundle,
	registryKind artifact.RegistryKind,
	finalizer ports.OperationFinalizer,
	airGap bool,
) (*DeliveryProcessor, error) {
	onlineValid := registry != nil && trust.Kind == registryKind
	if store == nil || clock == nil || finalizer == nil ||
		(!airGap && !onlineValid) || (airGap && registry != nil) {
		return nil, fmt.Errorf("platform authority delivery composition is invalid")
	}
	return &DeliveryProcessor{store: store, registry: registry, clock: clock,
		trust: trust, registryKind: registryKind, finalizer: finalizer, airGap: airGap}, nil
}

func (processor *DeliveryProcessor) Process(
	ctx context.Context,
	operationID string,
) error {
	if processor == nil || ctx == nil || operationID == "" {
		return ErrDeliveryPermanent
	}
	operation, err := processor.store.LoadOperation(ctx, operationID)
	if err != nil {
		return fmt.Errorf("load protected authority delivery: %w", err)
	}
	if processor.airGap && operation.Status != ports.StatusCompleted {
		return ErrDeliveryAwaitingImport
	}
	if operation.Status == ports.StatusProposalSigned {
		operation, err = processor.exchange(ctx, operation)
		if err != nil {
			return err
		}
	}
	if operation.Status == ports.StatusAckVerified {
		committed, commitErr := processor.store.CommitAcknowledged(ctx, operationID)
		if commitErr != nil {
			return fmt.Errorf("commit protected authority delivery: %w", commitErr)
		}
		operation = committed.Operation
	}
	if operation.Status != ports.StatusCompleted {
		return fmt.Errorf("%w: incomplete authority operation", ErrDeliveryPermanent)
	}
	return processor.finalize(ctx, operation)
}

func (processor *DeliveryProcessor) exchange(
	ctx context.Context,
	operation ports.OperationRecord,
) (ports.OperationRecord, error) {
	ackBytes, err := processor.registry.ExchangeProposal(ctx, operation.Proposal)
	if err != nil {
		return ports.OperationRecord{}, fmt.Errorf("exchange protected authority proposal: %w", err)
	}
	ack, err := artifact.VerifyAck(
		ackBytes, operation.Proposal, processor.trust, processor.clock.Now(),
	)
	if err != nil {
		return ports.OperationRecord{},
			fmt.Errorf("%w: verify authority acknowledgement", ErrDeliveryPermanent)
	}
	persisted, err := processor.store.PersistAcknowledgement(
		ctx, operation.OperationID, ack,
	)
	if err != nil {
		return ports.OperationRecord{}, fmt.Errorf("persist authority acknowledgement: %w", err)
	}
	return persisted, nil
}

func (processor *DeliveryProcessor) finalize(
	ctx context.Context,
	operation ports.OperationRecord,
) error {
	if operation.Purpose == "proposal_signer_rotation" {
		coordinator := &Coordinator{store: processor.store}
		if err := coordinator.applyCompletedRotation(ctx, operation); err != nil {
			return fmt.Errorf("finalize authority signer rotation: %w", err)
		}
		return nil
	}
	if operation.Purpose == "compensate_uncommitted_ack" {
		return nil
	}
	if err := processor.finalizer.FinalizeAuthorityOperation(ctx, operation); err != nil {
		return fmt.Errorf("finalize protected authority owner: %w", err)
	}
	return nil
}
