package config

import (
	"strings"
	"testing"
)

func TestFormsExportS3RejectsInvalidPersistentSettings(t *testing.T) {
	tests := []struct {
		name        string
		environment Environment
		replacement string
		remove      string
	}{
		{"production HTTP", EnvironmentProduction, formsExportEndpointKey + "=http://s3.internal:9000", ""},
		{"endpoint path", EnvironmentDevelopment, formsExportEndpointKey + "=https://s3.invalid/root", ""},
		{"region uppercase", EnvironmentDevelopment, formsExportRegionKey + "=EU-1", ""},
		{"bucket uppercase", EnvironmentDevelopment, formsExportBucketKey + "=Invalid-Bucket", ""},
		{"bucket IP", EnvironmentDevelopment, formsExportBucketKey + "=127.0.0.1", ""},
		{"owner width", EnvironmentDevelopment, formsExportOwnerKey + "=123", ""},
		{"path style invalid", EnvironmentDevelopment, formsExportPathStyleKey + "=yes", ""},
		{"missing path style", EnvironmentDevelopment, "", formsExportPathStyleKey},
		{"credential source invalid", EnvironmentDevelopment, formsExportCredentialSourceKey + "=ambient", ""},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := formsExportTestValues(t)
			if test.replacement != "" {
				key, value, _ := strings.Cut(test.replacement, "=")
				values[key] = value
			}
			delete(values, test.remove)
			if _, err := loadFormsExportS3(values, test.environment); err == nil {
				t.Fatal("invalid S3 config accepted")
			}
		})
	}
}

func TestFormsExportCredentialSourceIsFailClosed(t *testing.T) {
	values := formsExportTestValues(t)
	values[formsExportCredentialSourceKey] = string(FormsExportStaticFiles)
	if _, err := loadFormsExportS3Credentials(values); err == nil ||
		!strings.Contains(err.Error(), formsExportAccessKeyFileKey) {
		t.Fatalf("missing static files error = %v", err)
	}
	values = formsExportTestValues(t)
	configuredPath := t.TempDir()
	values[formsExportAccessKeyFileKey] = configuredPath
	if _, err := loadFormsExportS3Credentials(values); err == nil ||
		strings.Contains(err.Error(), configuredPath) {
		t.Fatalf("workload/static conflict error = %v", err)
	}
}

func formsExportTestValues(t *testing.T) environmentValues {
	t.Helper()
	values := make(environmentValues)
	for _, item := range validFormsExportEnvironment(false) {
		key, value, found := strings.Cut(item, "=")
		if !found {
			t.Fatal("invalid fixture entry")
		}
		values[key] = value
	}
	return values
}
