package compatibility

import compatibilityapp "github.com/Onurlulardan/harpy/core/internal/modules/analytics/application/compatibility"

func scanMapping(row row) (compatibilityapp.LegacyMapping, error) {
	result := compatibilityapp.LegacyMapping{}
	var publicID []byte
	err := row.Scan(&result.Key.TenantIdentity, &result.Key.Method, &result.Key.Path,
		&result.Key.Slug, &result.TenantID, &result.LegacyEndpointID, &publicID,
		&result.ReleaseID, &result.State, &result.HitCount, &result.LastHitAt,
		&result.ConsumersAckedAt, &result.RemoveAfter)
	if err != nil {
		return compatibilityapp.LegacyMapping{}, mapError(err)
	}
	if len(publicID) != 16 {
		return compatibilityapp.LegacyMapping{}, compatibilityapp.ErrUnavailable
	}
	copy(result.PublicID[:], publicID)
	return result, nil
}

func mappingArgs(value compatibilityapp.LegacyMapping) []any {
	return []any{value.Key.TenantIdentity, value.Key.Method, value.Key.Path,
		value.Key.Slug, value.TenantID, value.LegacyEndpointID,
		value.PublicID[:], value.ReleaseID}
}
