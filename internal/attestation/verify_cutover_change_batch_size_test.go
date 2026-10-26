package attestation

import (
	"bytes"
	"strings"
	"testing"

	cutoverv1 "github.com/Onurlulardan/harpy/contracts/cutover/v1"
)

func TestVerifyCutoverChangeBatchRejectsMaximumPlusOneByte(t *testing.T) {
	_, _, lease, _ := verifiedCutoverLeaseFixture(t)
	document := bytes.Repeat([]byte{' '}, cutoverv1.MaximumChangeBatchBytes+1)
	_, err := VerifyCutoverChangeBatch(document, lease)
	if err == nil || !strings.Contains(err.Error(), "size is invalid") {
		t.Fatalf("maximum plus one byte result = %v", err)
	}
}
