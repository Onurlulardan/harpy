package application

import (
	"context"
	"sync"

	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/domain/artifact"
	"github.com/Onurlulardan/harpy/core/internal/modules/platformauthority/ports"
)

type memoryStore struct {
	mutex              sync.Mutex
	operations         map[string]ports.OperationRecord
	heads              map[string]ports.ExternalHead
	bindings           map[string]ports.BindingRecord
	failCommitOnce     bool
	failActivationOnce bool
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		operations: map[string]ports.OperationRecord{},
		heads:      map[string]ports.ExternalHead{},
		bindings:   map[string]ports.BindingRecord{},
	}
}

func (store *memoryStore) LoadBinding(
	_ context.Context, bindingID string,
) (ports.BindingRecord, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	value, found := store.bindings[bindingID]
	if !found {
		return ports.BindingRecord{}, ports.ErrNotFound
	}
	return value, nil
}

func (store *memoryStore) SaveBinding(
	_ context.Context, value ports.BindingRecord, _ *artifact.VerifiedBindingProof,
) error {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if existing, found := store.bindings[value.BindingID]; found && existing != value {
		return ports.ErrConflict
	}
	store.bindings[value.BindingID] = value
	return nil
}

func (store *memoryStore) LoadOperation(
	_ context.Context, operationID string,
) (ports.OperationRecord, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	operation, exists := store.operations[operationID]
	if !exists {
		return ports.OperationRecord{}, ports.ErrNotFound
	}
	return operation, nil
}

func (store *memoryStore) PrepareOperation(
	_ context.Context, operation ports.OperationRecord,
) (ports.OperationRecord, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	if existing, exists := store.operations[operation.OperationID]; exists {
		if existing.Proposal.SHA256 != operation.Proposal.SHA256 {
			return ports.OperationRecord{}, ports.ErrConflict
		}
		return existing, nil
	}
	if binding, found := store.bindings[operation.Proposal.Value.RegistryBindingID]; found && (binding.SignerStatus != "active" ||
		binding.SignerKID != operation.Proposal.Value.SignerKID) {
		return ports.OperationRecord{}, ports.ErrConflict
	}
	store.operations[operation.OperationID] = operation
	return operation, nil
}

func (store *memoryStore) PersistAcknowledgement(
	_ context.Context, operationID string, ack artifact.VerifiedAck,
) (ports.OperationRecord, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	operation, exists := store.operations[operationID]
	if !exists {
		return ports.OperationRecord{}, ports.ErrNotFound
	}
	if operation.Status != ports.StatusProposalSigned {
		if operation.Ack.SHA256 != ack.SHA256 {
			return ports.OperationRecord{}, ports.ErrConflict
		}
		return operation, nil
	}
	operation.Ack = ack
	operation.Status = ports.StatusAckVerified
	store.operations[operationID] = operation
	return operation, nil
}

func (store *memoryStore) CommitAcknowledged(
	_ context.Context, operationID string,
) (ports.CommitResult, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	operation, exists := store.operations[operationID]
	if !exists {
		return ports.CommitResult{}, ports.ErrNotFound
	}
	if operation.Status == ports.StatusCompleted {
		return ports.CommitResult{Operation: operation, Head: store.headFor(operation)}, nil
	}
	if operation.Status != ports.StatusAckVerified {
		return ports.CommitResult{}, ports.ErrConflict
	}
	binding, found := store.bindings[operation.Proposal.Value.RegistryBindingID]
	if found && (binding.SignerStatus != "active" ||
		binding.SignerKID != operation.Proposal.Value.SignerKID) {
		return ports.CommitResult{}, ports.ErrConflict
	}
	if store.failCommitOnce {
		store.failCommitOnce = false
		return ports.CommitResult{}, context.DeadlineExceeded
	}
	key := operationHeadKey(operation)
	current, exists := store.heads[key]
	value := operation.Proposal.Value
	if (value.Generation == 1) != !exists ||
		(exists && (value.PreviousHeadSHA256 == nil ||
			*value.PreviousHeadSHA256 != current.AckSHA256 || value.Generation != current.Generation+1)) {
		return ports.CommitResult{}, ports.ErrStaleHead
	}
	head := ports.ExternalHead{
		Key: ports.HeadKey{RegistryBindingID: value.RegistryBindingID,
			ScopeKind: value.ScopeKind, ScopeID: value.ScopeID},
		Generation: value.Generation, StateSHA256: value.StateSHA256,
		AckSHA256:        operation.Ack.SHA256,
		RegistrySequence: operation.Ack.Value.RegistrySequence,
	}
	operation.Status = ports.StatusCompleted
	operation.CompletedAt = testNow
	store.operations[operationID] = operation
	store.heads[key] = head
	return ports.CommitResult{Operation: operation, Head: head}, nil
}

func (store *memoryStore) LoadHead(
	_ context.Context, key ports.HeadKey,
) (ports.ExternalHead, error) {
	store.mutex.Lock()
	defer store.mutex.Unlock()
	head, exists := store.heads[headKey(key)]
	if !exists {
		return ports.ExternalHead{}, ports.ErrNotFound
	}
	return head, nil
}

func (store *memoryStore) headFor(operation ports.OperationRecord) ports.ExternalHead {
	return store.heads[operationHeadKey(operation)]
}

func operationHeadKey(operation ports.OperationRecord) string {
	value := operation.Proposal.Value
	return headKey(ports.HeadKey{
		RegistryBindingID: value.RegistryBindingID,
		ScopeKind:         value.ScopeKind, ScopeID: value.ScopeID,
	})
}

func headKey(key ports.HeadKey) string {
	return key.RegistryBindingID + "\x00" + string(key.ScopeKind) + "\x00" + key.ScopeID
}
