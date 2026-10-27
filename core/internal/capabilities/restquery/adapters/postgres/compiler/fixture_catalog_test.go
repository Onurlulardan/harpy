package compiler

import (
	"crypto/sha256"
	"strings"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/Onurlulardan/harpy/core/internal/platform/sqlidentifier"
	"github.com/google/uuid"
)

const (
	fixtureTenantID  = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa"
	fixtureActorID   = "bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb"
	fixtureResource  = "cccccccc-cccc-4ccc-8ccc-cccccccccccc"
	fixtureVersion   = "dddddddd-dddd-4ddd-8ddd-dddddddddddd"
	fixtureIdentity  = "query-identity"
	fixtureMandatory = `{"policy":"mandatory"}`
)

func fixtureToken(testingT testing.TB, value string) sqlidentifier.Token {
	testingT.Helper()
	token, err := sqlidentifier.MintDynamic(value)
	if err != nil {
		testingT.Fatal(err)
	}
	return token
}

func fixtureStorage(testingT testing.TB, value string) api.ResourceStorage {
	testingT.Helper()
	storage, err := api.NewResourceStorage(fixtureToken(testingT, value))
	if err != nil {
		testingT.Fatal(err)
	}
	return storage
}

func fixtureReference(testingT testing.TB, owner string) api.CatalogReference {
	testingT.Helper()
	hash, err := api.NewCatalogHash(strings.Repeat("01", sha256.Size))
	if err != nil {
		testingT.Fatal(err)
	}
	reference, err := api.NewCatalogReference(
		owner, fixtureResource, fixtureVersion, hash,
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return reference
}

func fixtureCursorBinding(testingT testing.TB, owner string) cursorBinding {
	testingT.Helper()
	return cursorBinding{
		snapshot:      fixtureSnapshotBinding(testingT),
		authorization: sha256.Sum256([]byte("authorization")),
		mandatory:     []byte(fixtureMandatory), reference: fixtureReference(testingT, owner),
		identity: sha256.Sum256([]byte(fixtureIdentity)),
	}
}

func fixtureSnapshotBinding(testingT testing.TB) resolved.SnapshotBinding {
	testingT.Helper()
	binding, err := resolved.NewTenantSnapshotBinding(
		uuid.MustParse(fixtureTenantID), uuid.MustParse(fixtureActorID),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return binding
}

func fixtureBinding(
	testingT testing.TB, owner string, query resolved.RecordQuery,
) statementBinding {
	testingT.Helper()
	source := fixtureCursorBinding(testingT, owner)
	continuation, err := continuationIdentity(query)
	if err != nil {
		testingT.Fatal(err)
	}
	binding, err := newStatementBinding(source, continuation)
	if err != nil {
		testingT.Fatal(err)
	}
	return binding
}

func fixtureRecordInput(
	testingT testing.TB,
	owner string,
	query resolved.RecordQuery,
	visibility resolved.ResourceVisibility,
) recordInput {
	testingT.Helper()
	source := fixtureCursorBinding(testingT, owner)
	context, err := cursorContext(source, query)
	if err != nil {
		testingT.Fatal(err)
	}
	return recordInput{
		storage: fixtureStorage(testingT, "frm_records"), query: query,
		visibility: visibility, context: context,
		binding: fixtureBinding(testingT, owner, query),
	}
}

func fixturePath(testingT testing.TB, value string) resolved.RequestPath {
	testingT.Helper()
	path, err := resolved.NewRequestPath(value)
	if err != nil {
		testingT.Fatal(err)
	}
	return path
}

func fixtureAllVisibility(testingT testing.TB) resolved.ResourceVisibility {
	return fixtureAllVisibilityOwner(testingT, "forms")
}

func fixtureAllVisibilityOwner(
	testingT testing.TB, owner string,
) resolved.ResourceVisibility {
	testingT.Helper()
	visibility, err := resolved.NewAllResourceVisibility(
		sha256.Sum256([]byte("requirement")), fixtureReference(testingT, owner),
		fixturePath(testingT, ""),
	)
	if err != nil {
		testingT.Fatal(err)
	}
	return visibility
}
