//go:build integration

package config

func IntegrationLoadAPI(entries []string) (API, error) {
	return loadAPI(entries)
}
