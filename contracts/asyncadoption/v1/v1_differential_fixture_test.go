package asyncadoptionv1

import (
	"crypto/sha256"
	"embed"
	"encoding/hex"
	"testing"
	"time"

	asyncmanifestv1 "github.com/Onurlulardan/harpy/contracts/asyncmanifest/v1"
	eventsv1 "github.com/Onurlulardan/harpy/contracts/events/v1"
	jobsv1 "github.com/Onurlulardan/harpy/contracts/jobs/v1"
)

//go:embed testdata/v1-*.json
var v1DifferentialFixtures embed.FS

func TestV1QueueAndTimerFixturesMapToChecksummedJobs(t *testing.T) {
	cases := []struct {
		file, messageType, source, contract, target string
		class                                       InventoryClass
		surface                                     SourceSurface
	}{
		{"v1-form-event.json", "form.event", V1FormEventQueue,
			V1QueueContract, "workflow.handle_form_event", ClassQueue, SurfaceQueue},
		{"v1-schedule-event.json", "schedule.event", V1ScheduleQueue,
			V1QueueContract, "workflow.start_schedule", ClassQueue, SurfaceQueue},
		{"v1-delay-event.json", "workflow.delay.complete", V1DelayQueue,
			V1DelayContract, "workflow.resume_delay", ClassTimer, SurfaceDelayedPlugin},
	}
	for _, test := range cases {
		t.Run(test.messageType, func(t *testing.T) {
			document := readV1Fixture(t, test.file)
			message, err := DecodeV1QueueMessage(document)
			if err != nil {
				t.Fatal(err)
			}
			if message.Type != test.messageType || message.ID == "" ||
				message.Payload["tenant_slug"] != "acme" {
				t.Fatalf("V1 queue fixture drifted: %#v", message)
			}
			item := differentialItem(
				t, document, message.ID, test.source, test.contract,
				test.target, test.class, test.surface, message.PublishedAt,
			)
			if item.SourcePayloadSHA256 != item.TargetPayloadSHA256 {
				t.Fatal("V1 immutable payload evidence changed during mapping")
			}
			if err := ValidateTargetPolicy(
				item, differentialTargetCatalog(t, test.target, "workflow"),
			); err != nil {
				t.Fatal(err)
			}
			manifest, err := asyncmanifestv1.Load()
			if err != nil {
				t.Fatal(err)
			}
			if ValidateTargetPolicy(item, manifest.Jobs()) == nil {
				t.Fatal("unshipped V1 target bypassed the closed release catalog")
			}
			producers, _ := eventsv1.NewProducerCatalog(nil)
			if _, err := eventsv1.DecodeEnvelope(document, producers); err == nil {
				t.Fatal("V1 queue wire was accepted as a V2 integration event")
			}
		})
	}
}

func TestV1ImportAndIdempotencyFixturesHaveExplicitDisposition(t *testing.T) {
	importDocument := readV1Fixture(t, "v1-import-job.json")
	importJob, err := DecodeV1ImportJob(importDocument)
	if err != nil {
		t.Fatal(err)
	}
	if importJob.Status != "running" || importJob.ClientRequestID == "" ||
		importJob.LeaseOwner == "" || importJob.LastHeartbeatAt == "" {
		t.Fatalf("V1 import state drifted: %#v", importJob)
	}
	item := differentialItem(
		t, importDocument, importJob.ID, V1ImportFutureState, V1ImportContract,
		"forms.import.resume", ClassFuture, SurfaceFutureState,
		time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC),
	)
	if item.IdempotencyIdentity != V1ImportFutureState+"/"+importJob.ID {
		t.Fatal("V1 import idempotency identity is not stable")
	}
	if err := ValidateTargetPolicy(
		item, differentialTargetCatalog(t, "forms.import.resume", "forms"),
	); err != nil {
		t.Fatal(err)
	}
	idempotencyDocument := readV1Fixture(t, "v1-idempotency-complete.json")
	cached, err := DecodeV1IdempotencyResponse(idempotencyDocument)
	if err != nil {
		t.Fatal(err)
	}
	decision, err := ClassifyV1Idempotency(idempotencyDocument)
	if err != nil {
		t.Fatal(err)
	}
	if cached.State != "complete" || cached.Status != 201 || decision.ReuseInV2() ||
		decision.SourceContract() != V1IdempotencyContract ||
		decision.Disposition() != "invalidate_at_cutover" {
		t.Fatalf("V1 idempotency disposition drifted: %#v", cached)
	}
	pendingDocument := readV1Fixture(t, "v1-idempotency-pending.json")
	pending, err := ClassifyV1Idempotency(pendingDocument)
	if err != nil {
		t.Fatal(err)
	}
	if pending.ReuseInV2() || pending.Disposition() != "invalidate_at_cutover" {
		t.Fatal("V1 pending idempotency claim was not invalidated at cutover")
	}
}

func differentialItem(
	t *testing.T,
	document []byte,
	id, source, contract, target string,
	class InventoryClass,
	surface SourceSurface,
	runAt time.Time,
) Item {
	t.Helper()
	digest := sha256.Sum256(document)
	encoded := hex.EncodeToString(digest[:])
	item, err := NewItem(Item{
		InventoryClass: class, SourceSurface: surface, SourceName: source,
		SourceItemID: id, SourceContractVersion: contract, SourceRevision: 1,
		SourcePayloadSHA256: encoded, TransformVersion: "v1-async-differential.v1",
		IdempotencyIdentity: source + "/" + id, TargetKind: target,
		TargetPayloadRevision: 1, TargetPayloadReference: "v1-adoption/" + id,
		TargetPayloadSHA256: encoded, TargetRunAt: runAt.UTC().Format(time.RFC3339Nano),
		TargetPriority: 1, TargetMaximumAttempts: 8,
		TargetEffectClass: EffectDatabaseTransactional,
		CapturedAt:        "2026-07-15T12:00:00Z",
	})
	if err != nil {
		t.Fatal(err)
	}
	return item
}

func readV1Fixture(t *testing.T, name string) []byte {
	t.Helper()
	document, err := v1DifferentialFixtures.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return document
}

func differentialTargetCatalog(
	t *testing.T,
	kind string,
	domain string,
) jobsv1.Catalog {
	t.Helper()
	policy, err := jobsv1.NewKindPolicy(
		kind, 1, domain, jobsv1.EffectDatabaseTransactional, 8,
	)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := jobsv1.NewCatalog([]jobsv1.KindPolicy{policy})
	if err != nil {
		t.Fatal(err)
	}
	return catalog
}
