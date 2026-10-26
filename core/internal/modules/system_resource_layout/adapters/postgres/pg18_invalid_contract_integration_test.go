//go:build integration

package postgres

import (
	"context"
	"database/sql"
	"testing"

	layoutapi "github.com/Onurlulardan/harpy/core/internal/modules/system_resource_layout/api"
)

const invalidContractRevisionID = "92929292-9292-7292-8292-929292929292"

func assertMalformedLayoutContracts(
	t *testing.T,
	ctx context.Context,
	database *sql.DB,
) {
	t.Helper()
	for _, definition := range []string{
		`[]`,
		`{"contract":"harpy/system-resource-layout/v2","surfaces":{}}`,
	} {
		_, err := runLayoutRuntimeText(
			ctx, database, "layout_a", createSQL,
			"93939393-9393-4393-8393-939393939393",
			string(layoutapi.ResourceRole), definition, layoutDigest(definition),
			layoutNow(30),
		)
		assertLayoutSQLState(t, err, "23514")
	}
	createLayoutRevision(
		t, ctx, database, "layout_a", invalidContractRevisionID,
		layoutapi.ResourceRole, adapterDefinition,
	)
	arguments := layoutPublishArguments(
		invalidContractRevisionID, 1, adapterDefinition, layoutNow(31),
	)
	arguments[5] = `[]`
	arguments[6] = layoutDigest(`[]`)
	_, err := runLayoutRuntimeText(
		ctx, database, "layout_a", publishSQL, arguments...,
	)
	assertLayoutSQLState(t, err, "23514")
	arguments = layoutPublishArguments(
		invalidContractRevisionID, 1, adapterDefinition, layoutNow(31),
	)
	arguments[8] = `[]`
	arguments[9] = layoutDigest(`[]`)
	_, err = runLayoutRuntimeText(
		ctx, database, "layout_a", publishSQL, arguments...,
	)
	assertLayoutSQLState(t, err, "23514")
}
