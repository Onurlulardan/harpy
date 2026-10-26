package iam

import (
	"testing"

	systemtransfer "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/systemtransfer"
	"github.com/google/uuid"
)

func TestSystemTransferJobPayloadContainsOnlyResourceAndTransferID(t *testing.T) {
	id := uuid.MustParse("019be000-0000-7000-8000-000000000001")
	reference, err := systemtransfer.JobReference("users", id)
	digest, digestErr := systemtransfer.JobReferenceSHA256(reference)
	if err != nil || digestErr != nil {
		t.Fatalf("reference errors=%v/%v", err, digestErr)
	}
	for _, policy := range []struct {
		name  string
		value func() error
	}{
		{name: "export", value: func() error {
			payload, decodeErr := DecodeSystemTransferJobPayload(
				SystemExportJobPolicy(), reference, digest,
			)
			if payload.Resource != SystemTransferUsers || payload.TransferID != id {
				t.Fatalf("payload=%+v", payload)
			}
			return decodeErr
		}},
		{name: "import", value: func() error {
			_, decodeErr := DecodeSystemTransferJobPayload(
				SystemImportJobPolicy(), reference, digest,
			)
			return decodeErr
		}},
		{name: "export reconcile", value: func() error {
			_, decodeErr := DecodeSystemTransferJobPayload(
				SystemExportReconcileJobPolicy(), reference, digest,
			)
			return decodeErr
		}},
	} {
		t.Run(policy.name, func(t *testing.T) {
			if err := policy.value(); err != nil {
				t.Fatal(err)
			}
		})
	}
	if _, err = DecodeSystemTransferJobPayload(
		SystemImportJobPolicy(), reference, digest[:63]+"0",
	); err != ErrSystemTransferJobPolicy {
		t.Fatalf("tampered digest error=%v", err)
	}
}
