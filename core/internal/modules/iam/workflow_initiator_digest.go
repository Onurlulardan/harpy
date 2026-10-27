package iam

import (
	"crypto/sha256"
	"encoding/binary"

	workflowinitiator "github.com/Onurlulardan/harpy/core/internal/modules/iam/api/workflowinitiator"
)

func workflowInitiatorSHA256(evidence workflowinitiator.Evidence) [32]byte {
	document := []byte("harpy/iam/workflow-initiator-evidence/v1\x00")
	document = append(document, evidence.TenantID[:]...)
	document = append(document, evidence.ActorID[:]...)
	document = append(document, evidence.Permission...)
	document = binary.BigEndian.AppendUint64(document, evidence.AuthorizationVersion)
	document = binary.BigEndian.AppendUint64(document, evidence.AuthorizationClock)
	document = append(document, evidence.PermissionID[:]...)
	document = binary.BigEndian.AppendUint64(document, uint64(evidence.VerifiedAt.UnixMicro()))
	return sha256.Sum256(document)
}
