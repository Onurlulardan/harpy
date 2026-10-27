package catalog

import (
	"context"
	"errors"
	"reflect"
	"testing"

	platformtx "github.com/Onurlulardan/harpy/core/internal/platform/transaction"
	"github.com/google/uuid"
)

type catalogGenerationFixture struct {
	documentsCalls int
	documentsErr   error
}

func (*catalogGenerationFixture) ResolveTenantAnalyticsGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 11, 12, nil
}

func (*catalogGenerationFixture) ResolveTenantBillingGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 21, 22, nil
}

func (*catalogGenerationFixture) ResolveTenantCatalogGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 31, 32, nil
}

func (fixture *catalogGenerationFixture) ResolveTenantDocumentsGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	fixture.documentsCalls++
	return 41, 42, fixture.documentsErr
}

func (*catalogGenerationFixture) ResolveTenantIAMGenerations(
	context.Context, uuid.UUID,
) (int64, int64, int64, error) {
	return 51, 52, 53, nil
}

func (*catalogGenerationFixture) ResolveTenantWorkflowGenerations(
	context.Context, uuid.UUID,
) (int64, int64, error) {
	return 61, 62, nil
}

func TestCatalogResumeGenerationsIncludeExactDocumentsAuthority(t *testing.T) {
	fixture := &catalogGenerationFixture{}
	groups, domains, err := resolveResumeGenerations(
		context.Background(), fixture, uuid.New(),
	)
	if err != nil {
		t.Fatal(err)
	}
	if fixture.documentsCalls != 1 ||
		!reflect.DeepEqual(groupNames(groups), []string{
			"analytics", "billing", "catalog", "documents",
			"iam_system_resources", "workflow",
		}) ||
		!reflect.DeepEqual(domainNames(domains), []string{
			"analytics", "billing", "catalog", "documents",
			"iam", "system_resource_layout", "workflow",
		}) {
		t.Fatalf("Catalog resume authority drifted: groups=%v domains=%v",
			groups, domains)
	}
}

func TestCatalogLifecycleGenerationsDoNotWidenToDocuments(t *testing.T) {
	fixture := &catalogGenerationFixture{documentsErr: errors.New("unavailable")}
	groups, domains, err := resolveLifecycleGenerations(
		context.Background(), fixture, uuid.New(),
	)
	if err != nil || fixture.documentsCalls != 0 ||
		contains(groupNames(groups), "documents") ||
		contains(domainNames(domains), "documents") {
		t.Fatalf("Catalog lifecycle authority widened: groups=%v domains=%v err=%v",
			groups, domains, err)
	}
}

func groupNames(values []platformtx.CommandAuthorityGroup) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.ID
	}
	return result
}

func domainNames(values []platformtx.CommandAuthorityDomain) []string {
	result := make([]string, len(values))
	for index, value := range values {
		result[index] = value.Name
	}
	return result
}

func contains(values []string, expected string) bool {
	for _, value := range values {
		if value == expected {
			return true
		}
	}
	return false
}
