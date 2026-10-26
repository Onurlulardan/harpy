package worker

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	"github.com/Onurlulardan/harpy/core/internal/platform/admission"
	platformpostgres "github.com/Onurlulardan/harpy/core/internal/platform/postgres"
	"github.com/Onurlulardan/harpy/core/internal/platform/randomid"
	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

func buildIAMImportSourceCleanup(
	pool *platformpostgres.Pool,
	coordinator *platformtx.Coordinator,
	artifacts iammodule.SystemTransferArtifacts,
	gate *admission.Gate,
) (*iammodule.SystemTransferImportCleanupRuntime, error) {
	store, err := iammodule.BindSystemTransferImportCleanupStore(pool, coordinator)
	if err != nil {
		return nil, err
	}
	values := iamImportCleanupValues{}
	return iammodule.NewSystemTransferImportCleanupRuntime(
		store, artifacts, values, values, iamImportCleanupJitter{}, gate,
	)
}

type iamImportCleanupValues struct{}

func (iamImportCleanupValues) Now() time.Time {
	return time.Now().UTC().Truncate(time.Microsecond)
}

func (iamImportCleanupValues) NewV7(
	occurredAt time.Time,
) (iammodule.SystemTransferImportCleanupID, error) {
	value, err := randomid.NewUUIDv7(occurredAt)
	if err != nil {
		return iammodule.SystemTransferImportCleanupID{},
			fmt.Errorf("generate IAM import cleanup identity: %w", err)
	}
	return value, nil
}

type iamImportCleanupJitter struct{}

func (iamImportCleanupJitter) Uint64N(bound uint64) uint64 {
	if bound == 0 {
		panic("IAM import cleanup jitter bound is zero")
	}
	value, err := rand.Int(rand.Reader, new(big.Int).SetUint64(bound))
	if err != nil {
		panic("IAM import cleanup entropy is unavailable")
	}
	return value.Uint64()
}
