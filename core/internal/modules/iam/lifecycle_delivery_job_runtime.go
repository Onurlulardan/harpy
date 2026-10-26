package iam

import (
	"fmt"
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

type LifecycleDeliveryJobRuntime struct {
	provider *lifecycleDeliveryProvider
	closed   atomic.Bool
}

func NewLifecycleDeliveryJobRuntime(coordinator *platformtx.Coordinator,
	opener *gcpkms.DeliveryOpener,
	sender *smtp.LifecycleSender) (*LifecycleDeliveryJobRuntime, error) {
	if coordinator == nil || opener == nil || sender == nil {
		return nil, ErrLifecycleDeliveryRuntime
	}
	state, err := NewLifecycleDeliveryStateRuntime(coordinator)
	if err != nil {
		return nil, err
	}
	provider, err := newLifecycleDeliveryProvider(state,
		gcpLifecycleCredentialOpener{opener: opener}, sender, time.Now)
	if err != nil {
		return nil, err
	}
	return &LifecycleDeliveryJobRuntime{provider: provider}, nil
}

func (runtime *LifecycleDeliveryJobRuntime) Processors(catalog jobsv1.Catalog,
	store asyncpostgres.WorkerStore) ([]jobexecution.Processor, error) {
	if runtime == nil || runtime.closed.Load() || runtime.provider == nil || store == nil {
		return nil, ErrLifecycleDeliveryRuntime
	}
	policy, err := catalog.Resolve(LifecycleDeliveryJobKind, 1)
	if err != nil || policy.OwnerDomain != "iam" ||
		policy.EffectClass != jobsv1.EffectExternal || policy.MaximumAttempts != 8 {
		return nil, fmt.Errorf("IAM lifecycle delivery job policy is invalid")
	}
	boundPreEffect, err := asyncpostgres.BindExternalPreEffect(store)
	if err != nil {
		return nil, fmt.Errorf("bind IAM lifecycle delivery pre-effect: %w", err)
	}
	effects, err := effectexecution.NewRegistry(
		[]effectexecution.Provider{runtime.provider})
	if err != nil {
		return nil, fmt.Errorf("build IAM lifecycle provider registry: %w", err)
	}
	processor, err := jobexecution.NewExternalProcessor(policy,
		lifecycleDeliveryPlanner{}, invitationDeliveryPreEffect{store: boundPreEffect},
		store, effects)
	if err != nil {
		return nil, fmt.Errorf("build IAM lifecycle delivery processor: %w", err)
	}
	return []jobexecution.Processor{processor}, nil
}

func (runtime *LifecycleDeliveryJobRuntime) Close() error {
	if runtime != nil {
		runtime.closed.Store(true)
	}
	return nil
}
