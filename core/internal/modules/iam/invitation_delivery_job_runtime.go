package iam

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/gcpkms"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/smtp"
	asyncpostgres "github.com/Onurlulardan/harpy/core/internal/platform/async/adapters/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/effectexecution"
	"github.com/Onurlulardan/harpy/core/internal/platform/async/jobexecution"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type InvitationDeliveryJobRuntime struct {
	opener   *gcpkms.DeliveryOpener
	provider *invitationDeliveryProvider
	close    sync.Once
	closed   atomic.Bool
	closeErr error
}

func NewInvitationDeliveryJobRuntime(
	coordinator *platformtx.Coordinator,
	opener *gcpkms.DeliveryOpener,
	sender *smtp.Client,
) (*InvitationDeliveryJobRuntime, error) {
	if coordinator == nil || opener == nil || sender == nil {
		return nil, ErrInvitationDeliveryRuntime
	}
	state, err := NewInvitationDeliveryRuntime(coordinator)
	if err != nil {
		return nil, err
	}
	provider, err := newInvitationDeliveryProvider(
		state, gcpInvitationCredentialOpener{opener: opener}, sender, time.Now,
	)
	if err != nil {
		return nil, err
	}
	return &InvitationDeliveryJobRuntime{opener: opener, provider: provider}, nil
}

func (runtime *InvitationDeliveryJobRuntime) Processors(
	catalog jobsv1.Catalog,
	store asyncpostgres.WorkerStore,
) ([]jobexecution.Processor, error) {
	if runtime == nil || runtime.closed.Load() || runtime.provider == nil || store == nil {
		return nil, ErrInvitationDeliveryRuntime
	}
	policy, err := catalog.Resolve(InvitationDeliveryJobKind, 1)
	if err != nil || policy.OwnerDomain != "iam" ||
		policy.EffectClass != jobsv1.EffectExternal || policy.MaximumAttempts != 8 {
		return nil, fmt.Errorf("IAM invitation delivery job policy is invalid")
	}
	boundPreEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, fmt.Errorf("bind IAM invitation delivery pre-effect: %w", err)
	}
	effects, err := effectexecution.NewRegistry(
		[]effectexecution.Provider{runtime.provider},
	)
	if err != nil {
		return nil, fmt.Errorf("build IAM invitation delivery provider registry: %w", err)
	}
	processor, err := jobexecution.NewExternalProcessor(
		policy, invitationDeliveryPlanner{},
		invitationDeliveryPreEffect{store: boundPreEffect}, store, effects,
	)
	if err != nil {
		return nil, fmt.Errorf("build IAM invitation delivery processor: %w", err)
	}
	return []jobexecution.Processor{processor}, nil
}

func (runtime *InvitationDeliveryJobRuntime) Close() error {
	if runtime == nil {
		return nil
	}
	runtime.close.Do(func() {
		runtime.closed.Store(true)
		if runtime.opener == nil {
			runtime.closeErr = ErrInvitationDeliveryRuntime
			return
		}
		runtime.closeErr = errors.Join(runtime.closeErr, runtime.opener.Close())
	})
	return runtime.closeErr
}
