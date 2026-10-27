//go:build integration

package transaction

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func seedMultiGroupAuthorityRows(
	t *testing.T,
	ctx context.Context,
	admin *pgx.Conn,
	tenantID string,
) {
	t.Helper()
	statements := []string{
		"INSERT INTO platform.cutover_groups VALUES('tenant',$1,'billing','v2','v2',1)",
		"INSERT INTO platform.cutover_groups VALUES('tenant',$1,'workflow','v2','v2',1)",
		"INSERT INTO platform.cutover_group_members VALUES" +
			"('tenant',$1,'billing',1,'billing')",
		"INSERT INTO platform.cutover_group_members VALUES" +
			"('tenant',$1,'workflow',1,'workflow')",
		"INSERT INTO platform.domain_cutovers VALUES('tenant',$1,'billing','v2','v2',1)",
		"INSERT INTO platform.domain_cutovers VALUES('tenant',$1,'workflow','v2','v2',1)",
	}
	for _, statement := range statements {
		if _, err := admin.Exec(ctx, statement, tenantID); err != nil {
			t.Fatalf("seed multi-group authority: %v", err)
		}
	}
}
