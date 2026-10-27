package compiler

import (
	"crypto/sha256"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/api"
	"github.com/Onurlulardan/harpy/core/internal/capabilities/restquery/resolved"
	"github.com/google/uuid"
)

type includePlanFixture struct {
	plan     resolved.AuthorizedPlan
	catalogs includeCatalogFixture
}

type includePlanFixtureOptions struct {
	tenant       uuid.UUID
	count        api.CountMode
	withIncludes bool
	rootFilter   bool
}

func newIncludePlanFixture(t *testing.T) includePlanFixture {
	return newIncludePlanFixtureForTenant(t, uuid.MustParse(fixtureTenantID))
}

func newIncludePlanFixtureForTenant(
	t *testing.T, tenant uuid.UUID,
) includePlanFixture {
	return newIncludePlanFixtureWithOptions(t, includePlanFixtureOptions{
		tenant: tenant, count: api.CountModeNone, withIncludes: true,
	})
}

func newRootParityPlanFixture(
	t *testing.T, count api.CountMode, withIncludes bool,
) includePlanFixture {
	return newIncludePlanFixtureWithOptions(t, includePlanFixtureOptions{
		tenant: uuid.MustParse(fixtureTenantID), count: count,
		withIncludes: withIncludes, rootFilter: true,
	})
}

func newIncludePlanFixtureWithOptions(
	t *testing.T, options includePlanFixtureOptions,
) includePlanFixture {
	t.Helper()
	catalogs := newIncludeCatalogFixture(t)
	storageA := includeTargetStorage(t, catalogs.set, catalogs.targetA)
	storageB := includeTargetStorage(t, catalogs.set, catalogs.targetB)
	storageC := includeTargetStorage(t, catalogs.set, catalogs.targetC)
	child := includeResolvedNode(
		t, catalogs.relationC, catalogs.targetC, storageC,
		catalogs.cID, catalogs.cLabel, "/include/0/include/0", nil, nil,
	)
	where := includePredicate(
		t, catalogs.aLabel, includeInjectionValue, "/include/0/where",
	)
	first := includeResolvedNode(
		t, catalogs.relationA, catalogs.targetA, storageA,
		catalogs.aID, catalogs.aLabel, "/include/0", &where,
		[]resolved.Include{child},
	)
	second := includeResolvedNode(
		t, catalogs.relationB, catalogs.targetB, storageB,
		catalogs.bID, catalogs.bLabel, "/include/1", nil, nil,
	)
	var includes []resolved.Include
	if options.withIncludes {
		includes = []resolved.Include{first, second}
	}
	var rootWhere *resolved.Filter
	if options.rootFilter {
		filter := includePredicate(t, catalogs.rootLabel, "root-filter", "/where")
		rootWhere = &filter
	}
	page, _ := resolved.NewPage(2, nil)
	query, err := resolved.NewRecordQuery(
		[]resolved.Projection{
			includeProjection(t, catalogs.rootID, "id", "/select"),
			includeProjection(t, catalogs.rootLabel, "label", "/select/0"),
		}, rootWhere,
		[]resolved.Sort{
			includeSort(t, catalogs.rootLabel, api.DirectionAscending, "/sort/0", false),
			includeSort(t, catalogs.rootID, api.DirectionAscending, "/sort", true),
		}, nil, page, options.count, fixturePath(t, "/count"), true, includes,
	)
	if err != nil {
		t.Fatal(err)
	}
	rootStorage := includeTargetStorage(t, catalogs.set, catalogs.root)
	spec, err := resolved.NewRecordQuerySpec(
		catalogs.set, catalogs.root, rootStorage, catalogs.rootID,
		fixturePath(t, ""), query,
	)
	if err != nil {
		t.Fatal(err)
	}
	visibilities := []resolved.ResourceVisibility{
		includeAllVisibility(t, catalogs.root, "", "root"),
	}
	if options.withIncludes {
		visibilities = includeVisibilities(t, catalogs)
	}
	plan, err := resolved.MintAuthorizedPlan(
		spec, options.tenant, uuid.MustParse(fixtureActorID),
		sha256.Sum256([]byte("include-authorization")), visibilities,
	)
	if err != nil {
		t.Fatal(err)
	}
	return includePlanFixture{plan: plan, catalogs: catalogs}
}

func includeTargetStorage(
	t *testing.T, set api.CatalogSet, reference api.CatalogReference,
) api.ResourceStorage {
	t.Helper()
	catalog, found, err := set.Catalog(reference)
	if err != nil || !found {
		t.Fatal("include catalog missing")
	}
	storage, err := catalog.Storage()
	if err != nil {
		t.Fatal(err)
	}
	return storage
}
