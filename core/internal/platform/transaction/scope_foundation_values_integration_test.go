//go:build integration

package transaction

const scopeEnvironmentID = "77777777-7777-4777-8777-777777777777"
const scopeReleaseID = "88888888-8888-4888-8888-888888888888"
const scopeMigrationRunID = "99999999-9999-4999-8999-999999999999"
const scopeExtensionRunID = "12121212-1212-4212-8212-121212121212"

const scopeReceiptSeedSQL = `INSERT INTO platform.bootstrap_receipts
SELECT kind,run_id,$1,$2,'fresh',system_identifier::text,database.oid
FROM pg_catalog.pg_control_system(),pg_catalog.pg_database AS database,
(VALUES ('cluster_bootstrap'::text,'13131313-1313-4313-8313-131313131313'::uuid),
('extension_admin','12121212-1212-4212-8212-121212121212'),
('credential_cleanup','14141414-1414-4414-8414-141414141414')) AS receipt(kind,run_id)
WHERE database.datname=pg_catalog.current_database()`
