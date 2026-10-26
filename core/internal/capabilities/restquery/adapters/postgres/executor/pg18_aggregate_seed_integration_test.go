//go:build integration

package executor

import "testing"

func (gate *pg18Gate) prepareAggregateFixture(t *testing.T) {
	t.Helper()
	pg18ExecuteSetup(t, gate.ctx, gate.admin, []string{
		`CREATE TABLE tenant_a.frm_executor_aggregate(
fld_id uuid PRIMARY KEY,fld_status text,fld_status_folded text,fld_quantity bigint,
fld_amount numeric(12,6))`,
		`INSERT INTO tenant_a.frm_executor_aggregate
(fld_id,fld_status,fld_status_folded,fld_quantity,fld_amount) VALUES
('63000000-0000-4000-8000-000000000001','P_EVEN','p_even',NULL,0.000000),
('63000000-0000-4000-8000-000000000002','P_EVEN','p_even',NULL,0.000001),
('63000000-0000-4000-8000-000000000003','P_ODD','p_odd',NULL,0.000001),
('63000000-0000-4000-8000-000000000004','P_ODD','p_odd',NULL,0.000002),
('63000000-0000-4000-8000-000000000005','N_EVEN','n_even',NULL,0.000000),
('63000000-0000-4000-8000-000000000006','N_EVEN','n_even',NULL,-0.000001),
('63000000-0000-4000-8000-000000000007','N_ODD','n_odd',NULL,-0.000001),
('63000000-0000-4000-8000-000000000008','N_ODD','n_odd',NULL,-0.000002),
('63000000-0000-4000-8000-000000000009','SAFE','safe',1,NULL),
('63000000-0000-4000-8000-000000000010','OVER','over',9223372036854775807,NULL),
('63000000-0000-4000-8000-000000000011','OVER','over',1,NULL)`,
		`GRANT SELECT ON tenant_a.frm_executor_aggregate TO forgecrud_runtime`,
	})
}
