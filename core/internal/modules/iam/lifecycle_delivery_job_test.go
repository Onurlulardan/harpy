package iam

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
	"github.com/Onurlulardan/harpy/core/internal/modules/iam/application/lifecycle"
	platformasync "github.com/Onurlulardan/harpy/core/internal/platform/async"
	"github.com/google/uuid"
)

func TestLifecycleDeliveryJobCarriesOnlyOpaqueArtifactReference(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	tenantID := uuid.New()
	reference, err := lifecycledeliveryv1.NewReference(tenantID, uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	change := lifecycle.Change{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		CommandKey: "iam.password_reset.start", AggregateKind: "one_time_token",
		AggregateID: uuid.New(), ActorID: uuid.New(), Revision: 1,
		DeliveryArtifactRef: reference.String(),
	}
	sink := &LifecycleChangeSink{catalogs: catalogs, now: func() time.Time {
		return time.Unix(2_000_000_000, 0).UTC()
	}}
	job, present, err := sink.deliveryJob(change, tenantID, 3, 5)
	if err != nil || !present || job.PayloadReference != reference.String() {
		t.Fatalf("delivery job present=%t ref=%q err=%v", present, job.PayloadReference, err)
	}
	wire, err := json.Marshal(job)
	if err != nil {
		t.Fatal(err)
	}
	for _, forbidden := range []string{"password", "secret", "credential_hmac", "plaintext"} {
		if strings.Contains(strings.ToLower(string(wire)), forbidden) {
			t.Fatalf("delivery job leaked forbidden field %q: %s", forbidden, wire)
		}
	}
}

func TestLifecycleDeliveryJobRejectsCrossTenantReference(t *testing.T) {
	catalogs, err := platformasync.LoadWriteCatalogs()
	if err != nil {
		t.Fatal(err)
	}
	reference, _ := lifecycledeliveryv1.NewReference(uuid.New(), uuid.New())
	change := lifecycle.Change{
		OperationID: uuid.New(), RequestSHA256: [32]byte{1},
		CommandKey: "iam.password_reset.start", AggregateKind: "one_time_token",
		AggregateID: uuid.New(), ActorID: uuid.New(), Revision: 1,
		DeliveryArtifactRef: reference.String(),
	}
	sink := &LifecycleChangeSink{catalogs: catalogs, now: time.Now}
	if _, _, err := sink.deliveryJob(change, uuid.New(), 1, 1); err == nil {
		t.Fatal("cross-tenant lifecycle delivery reference was accepted")
	}
}
