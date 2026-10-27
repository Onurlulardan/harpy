package smtp

import (
	"bytes"
	"testing"

	lifecycledeliveryv1 "github.com/Onurlulardan/harpy/contracts/lifecycledelivery/v1"
)

func TestLifecycleTemplatesIncludeTenantQuery(t *testing.T) {
	credential := []byte(
		"abcdefghijklmnopqrstuv.ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopq",
	)
	cases := []struct {
		purpose lifecycledeliveryv1.Purpose
		route   string
	}{
		{lifecycledeliveryv1.PurposeEmailVerification, "/verify-email/"},
		{lifecycledeliveryv1.PurposePasswordReset, "/reset-password/"},
		{lifecycledeliveryv1.PurposePortalSetup, "/portal/setup/"},
		{lifecycledeliveryv1.PurposeInitialAdminSetup, "/initial-admin/setup/"},
	}
	for _, testCase := range cases {
		t.Run(string(testCase.purpose), func(t *testing.T) {
			template, err := newBodyTemplateFor(
				"Open {{LIFECYCLE_URL}}", lifecyclePlaceholder,
			)
			if err != nil {
				t.Fatal(err)
			}
			var output bytes.Buffer
			if _, err := template.writerRoute(
				"https://app.example.invalid", testCase.route, credential,
				"harpy",
			)(&output); err != nil {
				t.Fatal(err)
			}
			want := "Open https://app.example.invalid" + testCase.route +
				string(credential) + "?tenant=harpy"
			if output.String() != want {
				t.Fatalf("lifecycle URL = %q, want %q", output.String(), want)
			}
		})
	}
}
