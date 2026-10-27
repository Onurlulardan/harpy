//go:build integration

package postgres

import (
	"fmt"
	"strings"
)

type iamV1AdoptionTarget struct {
	sourceKey string
	table     string
	targetKey string
	record    map[string]any
}

func adoptedOrganizationID(tenant iamFoundationTenant) string {
	return adoptedID(tenant, 1)
}

func adoptedID(tenant iamFoundationTenant, ordinal int) string {
	prefix := "a"
	if tenant.schema == "t_iam_b" {
		prefix = "b"
	}
	return fmt.Sprintf("%s1000000-0000-4000-8000-%012d", prefix, ordinal)
}

func adoptedMarker(tenant iamFoundationTenant) string {
	if tenant.schema == "t_iam_b" {
		return "b"
	}
	return "a"
}

func adoptedTarget(
	category, sourceKey, table, targetKey string, record map[string]any,
) iamV1AdoptionTarget {
	return iamV1AdoptionTarget{
		sourceKey: category + ":" + sourceKey,
		table:     table, targetKey: targetKey, record: record,
	}
}

func adoptedTargetCount(tenant iamFoundationTenant) int {
	count := 0
	for _, category := range iamV1AdoptionCategories[:14] {
		count += len(adoptedTargets(tenant, category))
	}
	return count
}

func adoptedHash(value string) string {
	return `\x` + strings.Repeat(value, 32)
}

const adoptedAt = "2026-07-16T01:00:00Z"
