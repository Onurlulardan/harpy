package main

import (
	"os"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
	"github.com/Onurlulardan/harpy/core/internal/platform/envfile"
)

func main() {
	preserved := preservedProcessKeys()
	os.Clearenv()
	if len(os.Args) != 3 {
		os.Stderr.WriteString("profile and env file are required")
		os.Exit(1)
	}
	if _, err := envfile.Consume([]string{"--env-file", os.Args[2]}, os.Args[1]); err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
	for key, value := range preserved {
		if err := os.Setenv(key, value); err != nil {
			os.Stderr.WriteString(err.Error())
			os.Exit(1)
		}
	}
	var err error
	switch os.Args[1] {
	case "api":
		_, err = config.LoadAPI()
	case "worker":
		_, err = config.LoadWorker()
	case "operator_initialize":
		_, err = config.LoadOperatorInitialize()
	default:
		os.Stderr.WriteString("unknown profile")
		os.Exit(1)
	}
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Exit(1)
	}
}

func preservedProcessKeys() map[string]string {
	preserved := map[string]string{}
	for _, key := range []string{
		"HARPY_PROTECTED_AUTHORITY_ACK_TRUST_BUNDLE_PATH",
		"HARPY_RELEASE_TRUST_ROOT_PATH",
	} {
		if value := os.Getenv(key); value != "" {
			preserved[key] = value
		}
	}
	return preserved
}
