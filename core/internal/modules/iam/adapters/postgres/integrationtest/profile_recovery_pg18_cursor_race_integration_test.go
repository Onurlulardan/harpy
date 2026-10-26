//go:build integration

package integrationtest

import (
	"context"
	"fmt"

	iammodule "github.com/Onurlulardan/harpy/core/internal/modules/iam"
	profilebackfillpostgres "github.com/Onurlulardan/harpy/core/internal/modules/iam/adapters/postgres/profilebackfill"
	profileapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profile"
	profilebackfillapp "github.com/Onurlulardan/harpy/core/internal/modules/iam/application/profilebackfill"
	"github.com/google/uuid"
)

func commitRecoveryNonProfileUpdate(
	ctx context.Context, db publicationDatabase,
	resource profilebackfillapp.SourceResource,
) error {
	tx, err := db.admin.Begin(ctx)
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(context.Background()) }()
	if _, err = tx.Exec(ctx, "SET LOCAL ROLE harpy_owner"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx,
		"SELECT pg_catalog.set_config('search_path',$1,true)",
		db.tenant.schema+",platform,pg_catalog"); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, "SET LOCAL lock_timeout='750ms'"); err != nil {
		return err
	}
	tag, err := tx.Exec(ctx, `UPDATE iam_users
SET revision=revision+1,updated_at=pg_catalog.clock_timestamp()
WHERE id=$1 AND revision=$2`, resource.ID, resource.Revision)
	if err != nil || tag.RowsAffected() != 1 {
		return fmt.Errorf("non-profile update rows=%d: %w", tag.RowsAffected(), err)
	}
	return tx.Commit(ctx)
}

func classifyRecoveryCASInventory(
	ctx context.Context,
	contract profilebackfillpostgres.RecoveryContract,
	classifier *iammodule.ProfileRecoveryInventoryClassifier,
	resources []profilebackfillapp.SourceResource,
) (profilebackfillapp.RecoveryInventory, error) {
	schemas := map[uuid.UUID]profileapp.Schema{
		contract.Source.ID(): contract.Source,
		contract.Target.ID(): contract.Target,
	}
	if err := classifier.AppendBatch(ctx, schemas, resources); err != nil {
		return profilebackfillapp.RecoveryInventory{}, err
	}
	return classifier.Finish(ctx)
}
