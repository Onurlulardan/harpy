package transaction

import "github.com/google/uuid"

type globalRequestKind string

const (
	globalCommandRequest               globalRequestKind = "command"
	globalAuthenticatedSnapshotRequest globalRequestKind = "authenticated_snapshot"
	globalPublicSnapshotRequest        globalRequestKind = "public_snapshot"
)

type GlobalScopeRequest struct {
	command               GlobalCommandCapability
	authenticatedSnapshot GlobalSnapshotCapability
	publicSnapshot        PublicGlobalSnapshotCapability
	kind                  globalRequestKind
}

func NewGlobalCommandScopeRequest(
	capability GlobalCommandCapability,
) (GlobalScopeRequest, error) {
	request := GlobalScopeRequest{command: capability, kind: globalCommandRequest}
	if !request.valid() {
		return GlobalScopeRequest{}, ErrInvalidGlobalScopeRequest
	}
	return request, nil
}

func NewPublicGlobalSnapshotScopeRequest(
	capability PublicGlobalSnapshotCapability,
) (GlobalScopeRequest, error) {
	request := GlobalScopeRequest{
		publicSnapshot: capability, kind: globalPublicSnapshotRequest,
	}
	if !request.valid() {
		return GlobalScopeRequest{}, ErrInvalidGlobalScopeRequest
	}
	return request, nil
}

func NewGlobalSnapshotScopeRequest(
	capability GlobalSnapshotCapability,
) (GlobalScopeRequest, error) {
	request := GlobalScopeRequest{
		authenticatedSnapshot: capability,
		kind:                  globalAuthenticatedSnapshotRequest,
	}
	if !request.valid() {
		return GlobalScopeRequest{}, ErrInvalidGlobalScopeRequest
	}
	return request, nil
}

func (request GlobalScopeRequest) valid() bool {
	switch request.kind {
	case globalCommandRequest:
		return request.command.valid() && !request.authenticatedSnapshot.issued &&
			!request.publicSnapshot.issued
	case globalAuthenticatedSnapshotRequest:
		return request.authenticatedSnapshot.valid() && !request.command.issued &&
			!request.publicSnapshot.issued
	case globalPublicSnapshotRequest:
		return request.publicSnapshot.valid() && !request.command.issued &&
			!request.authenticatedSnapshot.issued
	default:
		return false
	}
}

func (request GlobalScopeRequest) actorID() uuid.UUID {
	if request.kind == globalCommandRequest {
		return request.command.actorID
	}
	if request.kind == globalAuthenticatedSnapshotRequest {
		return request.authenticatedSnapshot.actorID
	}
	return uuid.Nil
}

func (request GlobalScopeRequest) operationID() string {
	if request.kind == globalCommandRequest {
		return request.command.operationID
	}
	if request.kind == globalAuthenticatedSnapshotRequest {
		return request.authenticatedSnapshot.operationID
	}
	return string(request.publicSnapshot.operation)
}

func (request GlobalScopeRequest) mode() ScopeMode {
	if request.kind == globalCommandRequest {
		return Command
	}
	return SnapshotQuery
}

func (request GlobalScopeRequest) generations() (int64, int64) {
	if request.kind == globalCommandRequest {
		return request.command.groupGeneration, request.command.billingGeneration
	}
	if request.kind == globalAuthenticatedSnapshotRequest {
		return request.authenticatedSnapshot.groupGeneration,
			request.authenticatedSnapshot.billingGeneration
	}
	return request.publicSnapshot.groupGeneration, request.publicSnapshot.billingGeneration
}
