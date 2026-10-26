package profilebackfill

const recoveryInventoryCursorName = "iam_profile_recovery_inventory_v1"

type recoveryInventoryCursor struct {
	store *boundStore
	scope tenantScope
	name  string
}

func validRecoveryInventoryCursorName(value string) bool {
	return value == recoveryInventoryCursorName
}

func (cursor *recoveryInventoryCursor) usable(ctxValid bool) bool {
	return cursor != nil && cursor.store != nil && ctxValid &&
		validRecoveryInventoryCursorName(cursor.name)
}

var _ RecoveryInventoryCursor = (*recoveryInventoryCursor)(nil)
