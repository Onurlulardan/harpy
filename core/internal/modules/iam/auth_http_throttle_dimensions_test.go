package iam

import (
	"slices"
	"testing"
)

func TestAuthHTTPThrottleDimensionsMatchADR0066(t *testing.T) {
	tests := []struct {
		name       string
		dimensions map[string][]string
		expected   []string
	}{
		{
			name: "environment login",
			dimensions: loginEnvironmentThrottleDimensions(
				"tenant-a", "admin@example.com", "192.0.2.1",
			),
			expected: []string{"identifier", "identifier_ip", "ip", "tenant_slug"},
		},
		{
			name:       "tenant login",
			dimensions: loginTenantThrottleDimensions("admin@example.com", "192.0.2.1"),
			expected:   []string{"identifier", "identifier_ip", "ip"},
		},
		{
			name:       "refresh",
			dimensions: refreshThrottleDimensions("selector", "192.0.2.1"),
			expected:   []string{"ip", "selector"},
		},
	}
	for _, test := range tests {
		actual := make([]string, 0, len(test.dimensions))
		for dimension, parts := range test.dimensions {
			if len(parts) == 0 {
				t.Fatalf("%s dimension %s has no HMAC input", test.name, dimension)
			}
			actual = append(actual, dimension)
		}
		slices.Sort(actual)
		if !slices.Equal(actual, test.expected) {
			t.Fatalf("%s dimensions = %v, want %v", test.name, actual, test.expected)
		}
	}
}
