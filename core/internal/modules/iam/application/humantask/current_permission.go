package humantask

import (
	"context"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
)

type CurrentPermissionEvidence struct {
	PermissionID         [16]byte
	AuthorizationVersion int64
}

func (evidence CurrentPermissionEvidence) valid() bool {
	return evidence.PermissionID != [16]byte{} &&
		evidence.PermissionID[8]&0xc0 == 0x80 && evidence.AuthorizationVersion > 0
}

type CurrentPermissionPort interface {
	ResolveCurrentHumanPermission(
		context.Context,
		platformtx.TenantScope,
		[16]byte,
		string,
	) (CurrentPermissionEvidence, error)
}
