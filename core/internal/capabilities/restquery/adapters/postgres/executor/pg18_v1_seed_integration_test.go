//go:build integration

package executor

import (
	"context"
	"testing"
)

func (gate *pg18Gate) seedV1StableRows(t *testing.T) {
	t.Helper()
	_, err := gate.admin.Exec(gate.ctx, `INSERT INTO tenant_a.frm_executor_fixture
(fld_id,fld_name,fld_name_folded,fld_metadata) VALUES
('40000000-0000-4000-8000-000000000004','Tie','tie','{"tie":1}'),
('40000000-0000-4000-8000-000000000005','Tie','tie','{"tie":2}'),
('40000000-0000-4000-8000-000000000006','Zed','zed','{"last":true}')`)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		_, _ = gate.admin.Exec(context.Background(),
			`DELETE FROM tenant_a.frm_executor_fixture WHERE fld_id IN
('40000000-0000-4000-8000-000000000004',
'40000000-0000-4000-8000-000000000005',
'40000000-0000-4000-8000-000000000006')`)
	})
}
