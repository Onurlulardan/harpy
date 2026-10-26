package resolved

import (
	"crypto/sha256"
	"encoding/binary"
	"time"
)

func authorizedPlanSeal(plan AuthorizedPlan) [sha256.Size]byte {
	digest := sha256.New()
	_, _ = digest.Write([]byte("restquery_authorized_plan_v1\x00"))
	_, _ = digest.Write(plan.spec.stateSeal[:])
	_, _ = digest.Write([]byte{byte(plan.snapshot.scope), byte(plan.snapshot.audience)})
	_, _ = digest.Write(plan.snapshot.scopeID[:])
	_, _ = digest.Write(plan.snapshot.subjectID[:])
	writeAuthorizedPlanBytes(digest, []byte(plan.snapshot.operationID))
	writeAuthorizedPlanBytes(digest, []byte(plan.snapshot.snapshotAt.Format(time.RFC3339Nano)))
	_, _ = digest.Write(plan.authorizationContext[:])
	writeAuthorizedPlanBytes(digest, plan.mandatoryCanonical)
	var result [sha256.Size]byte
	copy(result[:], digest.Sum(nil))
	return result
}

type authorizedPlanDigest interface {
	Write([]byte) (int, error)
}

func writeAuthorizedPlanBytes(writer authorizedPlanDigest, value []byte) {
	var size [8]byte
	binary.BigEndian.PutUint64(size[:], uint64(len(value)))
	_, _ = writer.Write(size[:])
	_, _ = writer.Write(value)
}
