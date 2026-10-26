package profilebackfill

import _ "embed"

//go:embed sql/tenant/create.sql
var createSQL string

//go:embed sql/tenant/read.sql
var readSQL string

//go:embed sql/tenant/read_receipt.sql
var readReceiptSQL string

//go:embed sql/tenant/read_recovery_parent.sql
var readRecoveryParentSQL string

//go:embed sql/tenant/read_recovery_inventory.sql
var readRecoveryInventorySQL string

//go:embed sql/tenant/fetch_recovery_inventory.sql
var fetchRecoveryInventorySQL string

//go:embed sql/tenant/close_recovery_inventory.sql
var closeRecoveryInventorySQL string

//go:embed sql/tenant/create_recovery.sql
var createRecoverySQL string

//go:embed sql/tenant/read_recovery_receipt.sql
var readRecoveryReceiptSQL string
