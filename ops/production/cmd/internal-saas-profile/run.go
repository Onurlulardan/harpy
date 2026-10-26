package main

import (
	"flag"
	"fmt"
	"io"
	"time"
)

func runCLI(
	arguments []string,
	stdout io.Writer,
	stderr io.Writer,
	_ func() time.Time,
) int {
	flags := flag.NewFlagSet("internal-saas-profile", flag.ContinueOnError)
	flags.SetOutput(stderr)
	output := flags.String("output", "", "dedicated output directory")
	registryURL := flags.String("authority-registry-url", "", "SaaS authority registry URL")
	proposerKMSKeyID := flags.String("proposer-kms-key-id", "", "authority proposer KMS key ID")
	ackTrustBundle := flags.String("ack-trust-bundle-path", "", "container trust bundle path")
	workloadIdentity := flags.String("workload-identity-config-path", "",
		"optional container WIF config path")
	if err := flags.Parse(arguments); err != nil {
		return 2
	}
	if flags.NArg() != 0 || *output == "" {
		fmt.Fprintln(stderr, "usage: internal-saas-profile -output <directory> "+
			"-authority-registry-url <https-url> -proposer-kms-key-id <id> "+
			"-ack-trust-bundle-path <absolute-container-path>")
		return 2
	}
	result, err := ensureProfile(profileOptions{
		output: *output, registryURL: *registryURL,
		proposerKMSKeyID: *proposerKMSKeyID, ackTrustBundle: *ackTrustBundle,
		workloadIdentity: *workloadIdentity,
	})
	if err != nil {
		fmt.Fprintf(stderr, "internal-saas-profile: %v\n", err)
		return 1
	}
	state := "current"
	if result.Generated {
		state = "generated"
	}
	fmt.Fprintf(stdout, "internal SaaS profile %s: %s\n", state, result.Output)
	return 0
}
