package customerworker

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"time"

	tenantapi "github.com/Onurlulardan/harpy/core/internal/modules/tenancy/api"
	"github.com/google/uuid"
)

type Worker struct {
	claims      ClaimStore
	authority   AuthorityAcknowledger
	recorder    AuthorityRecorder
	finalizer   Finalizer
	leaseOwner  string
	randomness  io.Reader
	leasePeriod time.Duration
}

func New(
	claims ClaimStore,
	authority AuthorityAcknowledger,
	recorder AuthorityRecorder,
	finalizer Finalizer,
	leaseOwner string,
) (*Worker, error) {
	if claims == nil || authority == nil || recorder == nil || finalizer == nil ||
		leaseOwner == "" || len(leaseOwner) > 128 {
		return nil, ErrInvalid
	}
	return &Worker{claims: claims, authority: authority, recorder: recorder,
		finalizer: finalizer, leaseOwner: leaseOwner, randomness: rand.Reader,
		leasePeriod: 2 * time.Minute}, nil
}

func (worker *Worker) RunOnce(ctx context.Context) (bool, error) {
	request, err := worker.claimRequest()
	if err != nil {
		return false, err
	}
	claim, err := ClaimNext(ctx, worker.claims, request)
	if errors.Is(err, ErrNoWork) {
		return false, nil
	}
	if err != nil || !claim.Valid() {
		return false, fmt.Errorf("claim customer provisioning operation: %w", err)
	}
	revision := claim.Revision
	if claim.State == tenantapi.OnboardingAuthorityPending {
		receipt, ackErr := worker.authority.AcknowledgePrepared(
			ctx, tenantapi.UUIDString(claim.Plan.OperationID()),
		)
		if ackErr != nil {
			return true, fmt.Errorf("acknowledge customer authority: %w", ackErr)
		}
		revision, err = worker.recorder.RecordAuthority(ctx, claim, receipt)
		if err != nil {
			return true, fmt.Errorf("record customer authority: %w", err)
		}
	}
	if err := worker.finalizer.Finalize(ctx, claim, revision); err != nil {
		return true, fmt.Errorf("finalize customer provisioning: %w", err)
	}
	return true, nil
}

func (worker *Worker) claimRequest() (ClaimRequest, error) {
	if worker == nil || worker.randomness == nil || worker.leasePeriod < time.Second {
		return ClaimRequest{}, ErrUnavailable
	}
	attempt, err := uuid.NewRandomFromReader(worker.randomness)
	secret := make([]byte, 32)
	if err != nil {
		return ClaimRequest{}, ErrUnavailable
	}
	if _, err = io.ReadFull(worker.randomness, secret); err != nil {
		clear(secret)
		return ClaimRequest{}, ErrUnavailable
	}
	digest := sha256.Sum256(secret)
	clear(secret)
	return ClaimRequest{AttemptID: tenantapi.UUID(attempt), LeaseOwner: worker.leaseOwner,
		LeaseTokenSHA256: hex.EncodeToString(digest[:]),
		LeaseDuration:    worker.leasePeriod}, nil
}
