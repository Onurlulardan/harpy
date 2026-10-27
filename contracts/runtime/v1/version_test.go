package runtimev1

import (
	"encoding/json"
	"strings"
	"testing"
)

const testCommit = "0123456789abcdef0123456789abcdef01234567"

func TestNewVersionMarshalsCanonicalWire(t *testing.T) {
	t.Parallel()
	version, err := NewVersion(
		ServiceCoreAPI,
		"2.4.1-rc.1+build-7",
		testCommit,
		"2026-07-14T10:11:12Z",
		EnvironmentProduction,
	)
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(version)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"service":"core-api","version":"2.4.1-rc.1+build-7",` +
		`"commit":"` + testCommit + `","build_time":"2026-07-14T10:11:12Z",` +
		`"go_version":"go1.26.5","contract_version":1,"environment":"production"}`
	if string(encoded) != want {
		t.Fatalf("unexpected wire\nwant: %s\n got: %s", want, encoded)
	}
}

func TestNewVersionAllowsDevelopmentProvenance(t *testing.T) {
	t.Parallel()
	version, err := NewVersion(
		ServiceCoreWorker,
		"0.0.0-dev",
		"unknown",
		"2026-07-14T13:11:12+03:00",
		EnvironmentDevelopment,
	)
	if err != nil {
		t.Fatal(err)
	}
	if err := version.Validate(); err != nil {
		t.Fatal(err)
	}
}

func TestVersionMarshalRejectsSchemaViolation(t *testing.T) {
	t.Parallel()
	base, err := NewVersion(
		ServiceCoreAPI,
		"1.2.3",
		testCommit,
		"2026-07-14T10:11:12Z",
		EnvironmentProduction,
	)
	if err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*Version)
	}{
		{name: "service", mutate: func(value *Version) { value.Service = Service("api") }},
		{name: "version", mutate: func(value *Version) { value.Version = "01.2.3" }},
		{name: "long version", mutate: func(value *Version) { value.Version = strings.Repeat("1", 129) }},
		{name: "commit", mutate: func(value *Version) { value.Commit = strings.ToUpper(testCommit) }},
		{name: "unknown production commit", mutate: func(value *Version) { value.Commit = "unknown" }},
		{name: "non utc production time", mutate: func(value *Version) { value.BuildTime = "2026-07-14T13:11:12+03:00" }},
		{name: "go version", mutate: func(value *Version) { value.GoVersion = "go1.26.4" }},
		{name: "contract", mutate: func(value *Version) { value.ContractVersion = 2 }},
		{name: "environment", mutate: func(value *Version) { value.Environment = Environment("local") }},
	}
	for _, test := range tests {
		test := test
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			value := base
			test.mutate(&value)
			if _, err := json.Marshal(value); err == nil {
				t.Fatal("invalid version crossed the JSON boundary")
			}
		})
	}
}

func TestNewVersionAcceptsFortyAndSixtyFourCharacterCommits(t *testing.T) {
	t.Parallel()
	for _, commit := range []string{testCommit, testCommit + "89abcdef0123456789abcdef"} {
		if _, err := NewVersion(
			ServiceCoreMigrator,
			"1.0.0",
			commit,
			"2026-07-14T10:11:12Z",
			EnvironmentStaging,
		); err != nil {
			t.Fatalf("commit length %d: %v", len(commit), err)
		}
	}
}
