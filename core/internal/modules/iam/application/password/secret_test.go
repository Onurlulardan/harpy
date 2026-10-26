package password

import (
	"context"
	"errors"
	"testing"

	"golang.org/x/text/unicode/norm"
)

type rejectorFunc func(context.Context, []byte) (bool, error)

func (rejector rejectorFunc) Blocked(
	ctx context.Context,
	candidate []byte,
) (bool, error) {
	return rejector(ctx, candidate)
}

func TestNewSecretNormalizesNFCAndDestroysItsCopy(t *testing.T) {
	raw := []byte("e\u0301-password-value")
	var observed []byte
	common := rejectorFunc(func(_ context.Context, candidate []byte) (bool, error) {
		observed = append([]byte(nil), candidate...)
		return false, nil
	})
	policy := passwordPolicyForTest(t, common, allowRejector)
	secret, err := policy.NewSecret(context.Background(), raw, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !norm.NFC.IsNormal(observed) || string(raw) == string(observed) {
		t.Fatalf("candidate was not normalized: %q", observed)
	}
	retained := secret.value
	secret.Destroy()
	for _, value := range retained {
		if value != 0 {
			t.Fatal("destroy did not wipe the normalized password")
		}
	}
	if secret.usable() {
		t.Fatal("destroyed secret remained usable")
	}
}

func TestNewSecretEnforcesScalarAndRejectorPolicy(t *testing.T) {
	allow := rejectorFunc(func(context.Context, []byte) (bool, error) { return false, nil })
	policy := passwordPolicyForTest(t, allow, allow)
	for _, candidate := range [][]byte{
		[]byte("too-short"),
		[]byte(string(make([]rune, MaximumScalars+1))),
		{0xff},
	} {
		if _, err := policy.NewSecret(
			context.Background(), candidate, nil,
		); !errors.Is(err, ErrPasswordInvalid) {
			t.Fatalf("candidate %x error = %v", candidate, err)
		}
	}
	deny := rejectorFunc(func(context.Context, []byte) (bool, error) { return true, nil })
	deniedPolicy := passwordPolicyForTest(t, deny, allow)
	if _, err := deniedPolicy.NewSecret(
		context.Background(), []byte("long-enough-password"), nil,
	); !errors.Is(err, ErrPasswordRejected) {
		t.Fatalf("deny error = %v", err)
	}
	policyErr := errors.New("blocklist unavailable")
	failing := rejectorFunc(func(context.Context, []byte) (bool, error) { return false, policyErr })
	failingPolicy := passwordPolicyForTest(t, allow, failing)
	if _, err := failingPolicy.NewSecret(
		context.Background(), []byte("long-enough-password"), nil,
	); !errors.Is(err, ErrPasswordPolicyUnavailable) || !errors.Is(err, policyErr) {
		t.Fatalf("policy error = %v", err)
	}
}

func TestPolicyRejectsNormalizedContextTerms(t *testing.T) {
	policy := passwordPolicyForTest(t, allowRejector, allowRejector)
	if _, err := policy.NewSecret(
		context.Background(), []byte("safe-AcmeCorp-password"),
		[]string{"  ACMECORP  "},
	); !errors.Is(err, ErrPasswordRejected) {
		t.Fatalf("context rejection error = %v", err)
	}
	secret, err := policy.NewSecret(
		context.Background(), []byte("safe-unrelated-password"), []string{"Al"},
	)
	if err != nil {
		t.Fatal(err)
	}
	secret.Destroy()
}

func TestPolicyRequiresBothBlocklistSources(t *testing.T) {
	if _, err := NewPolicy(nil, allowRejector); !errors.Is(err, ErrPasswordPolicyUnavailable) {
		t.Fatalf("missing common blocklist error = %v", err)
	}
	if _, err := NewPolicy(allowRejector, nil); !errors.Is(err, ErrPasswordPolicyUnavailable) {
		t.Fatalf("missing compromised blocklist error = %v", err)
	}
}

var allowRejector = rejectorFunc(
	func(context.Context, []byte) (bool, error) { return false, nil },
)

func passwordPolicyForTest(t *testing.T, common, compromised Rejector) Policy {
	t.Helper()
	policy, err := NewPolicy(common, compromised)
	if err != nil {
		t.Fatal(err)
	}
	return policy
}

func TestVerificationSecretAllowsLegacyLengthButKeepsUpperBound(t *testing.T) {
	secret, err := NewVerificationSecret([]byte("short"))
	if err != nil {
		t.Fatal(err)
	}
	secret.Destroy()
	if _, err := NewVerificationSecret(
		[]byte(string(make([]rune, MaximumScalars+1))),
	); !errors.Is(err, ErrPasswordInvalid) {
		t.Fatalf("oversized verification secret error = %v", err)
	}
}
