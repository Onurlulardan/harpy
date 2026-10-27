package authentication

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)

func TestInitialPasswordMaterialIsCanonicalConsumeOnceAndWiped(t *testing.T) {
	source := bytes.Repeat([]byte{9}, initialSelectorBytes+2*initialSecretBytes)
	material, err := generateInitialPasswordMaterial(bytes.NewReader(source))
	if err != nil {
		t.Fatal(err)
	}
	var cookieView, csrfView []byte
	var presented *PresentedInitialPassword
	err = material.RevealOnce(func(cookie, csrf []byte) error {
		cookieView, csrfView = cookie, csrf
		var parseErr error
		presented, parseErr = ParseInitialPassword(cookie, csrf)
		return parseErr
	})
	if err != nil || presented == nil || len(cookieView) != initialCookieBytes ||
		len(csrfView) != initialCSRFWireBytes || material.active {
		t.Fatalf("invalid revealed material: err=%v", err)
	}
	defer presented.Destroy()
	if !bytes.Equal(cookieView, make([]byte, len(cookieView))) ||
		!bytes.Equal(csrfView, make([]byte, len(csrfView))) {
		t.Fatal("revealed wire buffers were not wiped")
	}
	if err := material.RevealOnce(func([]byte, []byte) error { return nil }); err == nil {
		t.Fatal("initial password material was revealed twice")
	}
}

func TestInitialPasswordHMACBindsEveryAuthorityDimension(t *testing.T) {
	binding := initialBinding{
		tenantID:          uuid.MustParse("10000000-0000-4000-8000-000000000001"),
		userID:            uuid.MustParse("20000000-0000-4000-8000-000000000001"),
		operationID:       uuid.MustParse("30000000-0000-4000-8000-000000000001"),
		credentialID:      uuid.MustParse("40000000-0000-4000-8000-000000000001"),
		credentialVersion: 1,
	}
	presented := &PresentedInitialPassword{
		selector: [16]byte{1}, secret: [32]byte{2}, csrf: [32]byte{3}, active: true,
	}
	secretKey, csrfKey := [32]byte{4}, [32]byte{5}
	secretHMAC := initialSecretHMAC(secretKey, binding.tenantID, binding.userID,
		binding.operationID, binding.credentialID, binding.credentialVersion,
		presented.selector, presented.secret)
	csrfHMAC := initialCSRFHMAC(csrfKey, presented.selector, presented.csrf)
	if !presented.matches(secretHMAC, csrfHMAC, secretKey, csrfKey, binding) {
		t.Fatal("correct initial password binding was rejected")
	}
	mutations := []initialBinding{binding, binding, binding, binding, binding}
	mutations[0].tenantID = uuid.New()
	mutations[1].userID = uuid.New()
	mutations[2].operationID = uuid.New()
	mutations[3].credentialID = uuid.New()
	mutations[4].credentialVersion++
	for index, changed := range mutations {
		if presented.matches(secretHMAC, csrfHMAC, secretKey, csrfKey, changed) {
			t.Fatalf("binding substitution %d was accepted", index)
		}
	}
}

func TestInitialPasswordRouteAndCompletionBindingsRejectSubstitution(t *testing.T) {
	tenantID, otherTenant := uuid.New(), uuid.New()
	selector, otherSelector := [16]byte{1}, [16]byte{2}
	if InitialPasswordRouteBinding(tenantID, selector) ==
		InitialPasswordRouteBinding(otherTenant, selector) ||
		InitialPasswordRouteBinding(tenantID, selector) ==
			InitialPasswordRouteBinding(tenantID, otherSelector) {
		t.Fatal("initial password route binding is not tenant-selector bound")
	}
	continuation := InitialPasswordContinuation{
		TenantID: tenantID, UserID: uuid.New(), InstallationOperationID: uuid.New(),
		CredentialID: uuid.New(), CredentialVersion: 1, Selector: selector,
	}
	changed := continuation
	changed.CredentialVersion++
	if initialCompletionRequestSHA256(continuation) == initialCompletionRequestSHA256(changed) {
		t.Fatal("completion request binding ignored credential version")
	}
}

func TestInitialPasswordConstantsAndRedactionAreExact(t *testing.T) {
	if InitialPasswordLifetime != 10*time.Minute || InitialPasswordMaxAttempts != 5 {
		t.Fatalf("unexpected continuation limits: %s/%d",
			InitialPasswordLifetime, InitialPasswordMaxAttempts)
	}
	challenge := InitialPasswordChallenge{}
	for _, rendered := range []string{
		fmt.Sprint(challenge), fmt.Sprintf("%#v", challenge), challenge.LogValue().String(),
	} {
		if !strings.Contains(rendered, "REDACTED") {
			t.Fatalf("challenge was not redacted: %q", rendered)
		}
	}
}

func TestInitialPasswordOperationsFailClosedAndDestroySecretsOutsideScope(t *testing.T) {
	temporary := &trackedPasswordSecret{}
	if _, err := (InitialPasswordAuthenticator{}).Claim(
		context.Background(), CredentialScope{}, temporary,
	); err == nil || !temporary.destroyed {
		t.Fatalf("invalid start err=%v destroyed=%t", err, temporary.destroyed)
	}
	newSecret := &trackedPasswordSecret{}
	if _, err := (InitialPasswordAuthenticator{}).Finalize(
		context.Background(), CredentialScope{}, nil, newSecret, DeviceEvidence{},
	); err == nil || !newSecret.destroyed {
		t.Fatalf("invalid complete err=%v destroyed=%t", err, newSecret.destroyed)
	}
}
