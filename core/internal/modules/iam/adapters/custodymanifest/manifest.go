package custodymanifest

import (
	"fmt"
	"slices"

	iamrestorev1 "github.com/Onurlulardan/harpy/contracts/iamrestore/v1"
	"github.com/google/uuid"
)

type Manifest struct {
	value iamrestorev1.CustodyManifest
}

type Result struct {
	OperationID        uuid.UUID
	Generation         int64
	SecurityEpoch      [32]byte
	ResultSHA256       string
	PublicKIDs         []string
	RevokedGenerations []int64
}

func Parse(document []byte, environmentID uuid.UUID) (Manifest, error) {
	value, err := iamrestorev1.ParseCustodyManifest(document)
	if err != nil || environmentID == uuid.Nil || value.EnvironmentID != environmentID.String() {
		return Manifest{}, fmt.Errorf("IAM custody manifest binding is invalid")
	}
	return Manifest{value: value}, nil
}

// RotateForRestore consumes the idempotent public result projected by the
// custody provider. The provider performs secret rotation before publishing it.
func (manifest Manifest) RotateForRestore(operationID uuid.UUID) (Result, error) {
	if operationID == uuid.Nil {
		return Result{}, fmt.Errorf("IAM custody rotation operation is invalid")
	}
	needle := operationID.String()
	index, found := slices.BinarySearchFunc(manifest.value.Rotations, needle,
		func(value iamrestorev1.RotationResult, target string) int {
			if value.OperationID < target {
				return -1
			}
			if value.OperationID > target {
				return 1
			}
			return 0
		})
	if !found {
		return Result{}, fmt.Errorf("IAM custody rotation result is unavailable")
	}
	value := manifest.value.Rotations[index]
	epoch, err := decodeSecurityEpoch(value.SecurityEpoch)
	if err != nil {
		return Result{}, err
	}
	return Result{OperationID: operationID, Generation: value.Generation,
		SecurityEpoch: epoch, ResultSHA256: value.ResultSHA256,
		PublicKIDs:         append([]string(nil), value.PublicKIDs...),
		RevokedGenerations: append([]int64(nil), value.RevokedGenerations...)}, nil
}
