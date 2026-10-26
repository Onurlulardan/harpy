package iam

import (
	"context"
	"encoding/json"
	"errors"

	iampostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres"
	workprincipalapi "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workprincipal"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

var ErrWorkPrincipalEvidence = errors.New("IAM Work principal evidence unavailable")

type WorkPrincipalRuntime struct {
	reader iampostgres.WorkPrincipalReader
}

func NewWorkPrincipalRuntime() (*WorkPrincipalRuntime, error) {
	reader, err := iampostgres.NewWorkPrincipalReader()
	if err != nil {
		return nil, ErrWorkPrincipalEvidence
	}
	return &WorkPrincipalRuntime{reader: reader}, nil
}

func (runtime *WorkPrincipalRuntime) AvailableUsers(ctx context.Context,
	scope platformtx.TenantScope, search string, limit int,
) (json.RawMessage, error) {
	if runtime == nil || runtime.reader == nil {
		return nil, ErrWorkPrincipalEvidence
	}
	document, err := runtime.reader.AvailableUsers(ctx, scope, search, limit)
	if err != nil {
		return nil, ErrWorkPrincipalEvidence
	}
	return document, nil
}

func (runtime *WorkPrincipalRuntime) Evidence(ctx context.Context,
	scope platformtx.TenantScope, targetKind string, targetID [16]byte,
) (workprincipalapi.Evidence, error) {
	if runtime == nil || runtime.reader == nil {
		return workprincipalapi.Evidence{}, ErrWorkPrincipalEvidence
	}
	evidence, err := runtime.reader.Evidence(ctx, scope, targetKind, uuid.UUID(targetID))
	if err != nil {
		return workprincipalapi.Evidence{}, ErrWorkPrincipalEvidence
	}
	result := workprincipalapi.Evidence{AuthorizationRevision: evidence.AuthorizationClock,
		TargetExists: evidence.TargetExists, TargetRevision: evidence.TargetRevision,
		Principals: make([]workprincipalapi.Principal, 0, len(evidence.Principals))}
	for _, principal := range evidence.Principals {
		result.Principals = append(result.Principals, workprincipalapi.Principal{
			Kind: principal.Kind, ID: [16]byte(principal.ID), Revision: principal.Revision})
	}
	return result, nil
}

var _ workprincipalapi.Directory = (*WorkPrincipalRuntime)(nil)
