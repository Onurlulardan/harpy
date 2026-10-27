//go:build integration

package postgres

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	"github.com/google/uuid"
)

type terminalJobRoot struct {
	id, sourceID                         uuid.UUID
	payload, sourceSHA, payloadSHA       []byte
	reasonClass, reasonCode, writer      string
	groupGeneration, domainGeneration    int64
	recordedAt, fullUntil, metadataUntil time.Time
}

type terminalJobReference struct {
	Attempt          int    `json:"attempt"`
	ContractVersion  string `json:"contract_version"`
	JobID            string `json:"job_id"`
	Kind             string `json:"kind"`
	PayloadReference string `json:"payload_reference"`
	PayloadRevision  int    `json:"payload_revision"`
	PayloadSHA256    string `json:"payload_sha256"`
	ReasonCode       string `json:"reason_code"`
}

func assertTerminalJobRoot(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
	wantClass string,
	wantCode string,
) uuid.UUID {
	t.Helper()
	var root terminalJobRoot
	var payloadSHA, sourceSHA string
	err := fixture.owner.QueryRowContext(ctx, `
SELECT id,source_id,payload_canonical,payload_sha256,source_sha256,
       reason_class,reason_code,writer_origin,authority_group_generation,
       cutover_generation,recorded_at,full_payload_expires_at,metadata_expires_at
FROM platform.dead_letters
WHERE source_kind='job' AND record_kind='root' AND source_id=$1`, jobID).Scan(
		&root.id, &root.sourceID, &root.payload, &payloadSHA, &sourceSHA,
		&root.reasonClass, &root.reasonCode, &root.writer,
		&root.groupGeneration, &root.domainGeneration,
		&root.recordedAt, &root.fullUntil, &root.metadataUntil,
	)
	if err != nil {
		t.Fatalf("read terminal job dead-letter: %v", err)
	}
	root.payloadSHA = []byte(payloadSHA)
	root.sourceSHA = []byte(sourceSHA)
	canonical, err := canonicaljsonv1.Canonicalize(root.payload, 65_536)
	if err != nil || !bytes.Equal(canonical, root.payload) {
		t.Fatalf("terminal job payload is not RFC8785 canonical: %v", err)
	}
	digest := sha256.Sum256(root.payload)
	if hex.EncodeToString(digest[:]) != payloadSHA {
		t.Fatal("terminal job payload digest differs from canonical bytes")
	}
	var reference terminalJobReference
	if err := json.Unmarshal(root.payload, &reference); err != nil {
		t.Fatal(err)
	}
	assertTerminalRootIdentity(t, root, reference, jobID, wantClass, wantCode)
	return root.id
}

func assertTerminalRootIdentity(
	t *testing.T,
	root terminalJobRoot,
	reference terminalJobReference,
	jobID uuid.UUID,
	wantClass string,
	wantCode string,
) {
	t.Helper()
	wantID := terminalJobRootID(jobID, reference.Attempt, wantCode)
	if root.id != wantID || root.id.Version() != uuid.Version(8) ||
		root.id.Variant() != uuid.RFC4122 || root.sourceID != jobID {
		t.Fatalf("terminal root identity=%s want=%s", root.id, wantID)
	}
	if root.reasonClass != wantClass || root.reasonCode != wantCode ||
		root.writer != "v2" || root.groupGeneration < 1 || root.domainGeneration < 1 {
		t.Fatalf("terminal root disposition=%#v", root)
	}
	if reference.ContractVersion != "harpy.job-terminal-reference.v1" ||
		reference.JobID != jobID.String() || reference.Kind != "workflow.execute" ||
		reference.PayloadRevision != 1 || reference.PayloadSHA256 != string(root.sourceSHA) ||
		reference.ReasonCode != wantCode || reference.PayloadReference == "" {
		t.Fatalf("terminal job reference=%#v", reference)
	}
	if !root.fullUntil.Equal(root.recordedAt.Add(30*24*time.Hour)) ||
		!root.metadataUntil.Equal(root.recordedAt.Add(365*24*time.Hour)) {
		t.Fatal("terminal root retention does not use its terminal DB timestamp")
	}
}

func terminalJobRootID(jobID uuid.UUID, attempt int, reason string) uuid.UUID {
	document := fmt.Sprintf(
		"harpy.job-terminal-dead-letter.v1\njob_id=%s\nattempt=%d\nreason=%s",
		jobID, attempt, reason,
	)
	digest := sha256.Sum256([]byte(document))
	var result uuid.UUID
	copy(result[:], digest[:16])
	result[6] = result[6]&0x0f | 0x80
	result[8] = result[8]&0x3f | 0x80
	return result
}

func terminalJobRootCount(
	t *testing.T,
	ctx context.Context,
	fixture jobLifecyclePGFixture,
	jobID uuid.UUID,
) int {
	t.Helper()
	var count int
	if err := fixture.owner.QueryRowContext(ctx, `
SELECT count(*) FROM platform.dead_letters
WHERE source_kind='job' AND record_kind='root' AND source_id=$1`, jobID).Scan(&count); err != nil {
		t.Fatal(err)
	}
	return count
}
