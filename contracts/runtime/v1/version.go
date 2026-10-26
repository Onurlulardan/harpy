package runtimev1

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

type Version struct {
	Service         Service     `json:"service"`
	Version         string      `json:"version"`
	Commit          string      `json:"commit"`
	BuildTime       string      `json:"build_time"`
	GoVersion       string      `json:"go_version"`
	ContractVersion int         `json:"contract_version"`
	Environment     Environment `json:"environment"`
}

func NewVersion(
	service Service,
	releaseVersion string,
	commit string,
	buildTime string,
	environment Environment,
) (Version, error) {
	version := Version{
		Service:         service,
		Version:         releaseVersion,
		Commit:          commit,
		BuildTime:       buildTime,
		GoVersion:       GoVersion,
		ContractVersion: ContractVersion,
		Environment:     environment,
	}
	if err := version.Validate(); err != nil {
		return Version{}, err
	}
	return version, nil
}

func (version Version) Validate() error {
	if !version.Service.Valid() {
		return fmt.Errorf("version service is invalid")
	}
	if !validReleaseVersion(version.Version) {
		return fmt.Errorf("release version is invalid")
	}
	if !validCommit(version.Commit, version.Environment) {
		return fmt.Errorf("commit is invalid")
	}
	if !validBuildTime(version.BuildTime, version.Environment) {
		return fmt.Errorf("build time is invalid")
	}
	if version.GoVersion != GoVersion {
		return fmt.Errorf("Go version is invalid")
	}
	if version.ContractVersion != ContractVersion {
		return fmt.Errorf("contract version is invalid")
	}
	if !version.Environment.Valid() {
		return fmt.Errorf("environment is invalid")
	}
	return nil
}

func (version Version) MarshalJSON() ([]byte, error) {
	if err := version.Validate(); err != nil {
		return nil, err
	}
	type wireVersion Version
	return json.Marshal(wireVersion(version))
}

func validCommit(value string, environment Environment) bool {
	if value == "unknown" {
		return environment == EnvironmentDevelopment || environment == EnvironmentTest
	}
	if len(value) != 40 && len(value) != 64 {
		return false
	}
	for index := 0; index < len(value); index++ {
		character := value[index]
		if (character < '0' || character > '9') &&
			(character < 'a' || character > 'f') {
			return false
		}
	}
	return true
}

func validBuildTime(value string, environment Environment) bool {
	parsed, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return false
	}
	if environment != EnvironmentProduction && environment != EnvironmentStaging {
		return true
	}
	_, offset := parsed.Zone()
	return strings.HasSuffix(value, "Z") && offset == 0
}

func validReleaseVersion(value string) bool {
	if !validTextLength(value, 1, 128) {
		return false
	}
	coreAndPreRelease, build, hasBuild := strings.Cut(value, "+")
	if hasBuild && (!validVersionIdentifiers(build) || strings.Contains(build, "+")) {
		return false
	}
	core, preRelease, hasPreRelease := strings.Cut(coreAndPreRelease, "-")
	if hasPreRelease && !validVersionIdentifiers(preRelease) {
		return false
	}
	parts := strings.Split(core, ".")
	if len(parts) != 3 {
		return false
	}
	for _, part := range parts {
		if !validVersionNumber(part) {
			return false
		}
	}
	return true
}

func validVersionNumber(value string) bool {
	if value == "" || len(value) > 1 && value[0] == '0' {
		return false
	}
	for index := 0; index < len(value); index++ {
		if value[index] < '0' || value[index] > '9' {
			return false
		}
	}
	return true
}

func validVersionIdentifiers(value string) bool {
	for _, identifier := range strings.Split(value, ".") {
		if identifier == "" {
			return false
		}
		for index := 0; index < len(identifier); index++ {
			character := identifier[index]
			if !asciiAlphaNumeric(character) && character != '-' {
				return false
			}
		}
	}
	return true
}
