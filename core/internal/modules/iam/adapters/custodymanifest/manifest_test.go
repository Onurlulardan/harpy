package custodymanifest

import (
	"encoding/base64"
	"encoding/json"
	"reflect"
	"testing"

	canonicaljsonv1 "github.com/Onurlulardan/harpy/contracts/canonicaljson/v1"
	iamrestorev1 "github.com/Onurlulardan/harpy/contracts/iamrestore/v1"
	"github.com/google/uuid"
)

func TestRotateForRestoreIsExactAndIdempotent(t *testing.T) {
	environmentID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	operationID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	securityEpoch := [32]byte{9}
	rotation := iamrestorev1.RotationResult{OperationID: operationID.String(),
		Status: "rotated", Generation: 2,
		SecurityEpoch:      base64.RawURLEncoding.EncodeToString(securityEpoch[:]),
		PublicKIDs:         []string{base64.RawURLEncoding.EncodeToString(make([]byte, 16))},
		RevokedGenerations: []int64{1}}
	rotation.ResultSHA256 = iamrestorev1.RotationDigest(rotation)
	value := iamrestorev1.CustodyManifest{Schema: iamrestorev1.CustodyManifestSchema,
		EnvironmentID: environmentID.String(), Provider: "test.v1",
		KeyringHandle: "projects/test/keyrings/iam",
		Rotations:     []iamrestorev1.RotationResult{rotation}}
	document, _ := json.Marshal(value)
	document, _ = canonicaljsonv1.CanonicalizeObject(document, iamrestorev1.MaximumDocumentBytes)
	manifest, err := Parse(document, environmentID)
	if err != nil {
		t.Fatal(err)
	}
	first, err := manifest.RotateForRestore(operationID)
	if err != nil {
		t.Fatal(err)
	}
	second, err := manifest.RotateForRestore(operationID)
	if err != nil || !reflect.DeepEqual(first, second) || first.SecurityEpoch != securityEpoch {
		t.Fatalf("rotation replay differs: %#v %#v %v", first, second, err)
	}
	if _, err = manifest.RotateForRestore(uuid.New()); err == nil {
		t.Fatal("unstaged rotation was accepted")
	}
}
