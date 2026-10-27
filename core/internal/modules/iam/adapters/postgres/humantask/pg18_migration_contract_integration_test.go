//go:build integration

package humantask

import (
	"context"
	"testing"
	"time"
)

func TestIAMHumanTaskMigrationsAgainstPostgreSQL184(t *testing.T) {
	for _, origin := range []string{"fresh", "v1_adopted"} {
		t.Run(origin, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 8*time.Minute)
			defer cancel()
			database := startIAMHumanPG18(t, ctx, origin)
			prepareIAMHumanPG18(t, ctx, database, origin)
			for _, schema := range []string{"iam_human_a", "iam_human_b"} {
				var resourceID string
				var actions, permissions int
				err := database.QueryRowContext(ctx, `SELECT resource.id::text,
(SELECT count(*) FROM `+schema+`.iam_actions WHERE action_key IN
 ('claim','decide','reassign','expire','cancel')),
(SELECT count(*) FROM `+schema+`.iam_permissions AS permission
 JOIN `+schema+`.iam_resources AS resource ON resource.id=permission.resource_id
 WHERE resource.resource_key='workflow.human_tasks')
FROM `+schema+`.iam_resources AS resource
WHERE resource.resource_key='workflow.human_tasks'`).Scan(
					&resourceID, &actions, &permissions)
				if err != nil || actions != 5 || permissions != 5 {
					t.Fatalf("%s catalog=%s/%d/%d: %v", schema,
						resourceID, actions, permissions, err)
				}
				if origin == "v1_adopted" && resourceID !=
					"10000000-0000-4000-8000-000000000001" {
					t.Fatalf("%s adopted resource changed to %s", schema, resourceID)
				}
			}
		})
	}
}
