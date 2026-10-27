package api

import (
	"context"
	"testing"

	"github.com/Onurlulardan/harpy/core/internal/platform/config"
)

type iamDecryptorFixture struct{}

func (iamDecryptorFixture) DecryptTOTP(
	context.Context,
	string,
	[]byte,
	[]byte,
	[]byte,
	[32]byte,
) ([32]byte, error) {
	return [32]byte{}, nil
}

func (iamDecryptorFixture) Close() error { return nil }

func TestOpenIAMTOTPDecryptorPassesCanonicalConfiguredVersion(t *testing.T) {
	setAPIStartupEnvironment(t)
	settings, err := config.LoadAPI()
	if err != nil {
		t.Fatal(err)
	}
	iamSettings, err := settings.IAM()
	if err != nil {
		t.Fatal(err)
	}
	var received string
	decryptor, err := openIAMTOTPDecryptor(
		context.Background(), iamSettings,
		func(_ context.Context, keyID string) (iamTOTPDecryptor, error) {
			received = keyID
			return iamDecryptorFixture{}, nil
		},
	)
	if err != nil || decryptor == nil {
		t.Fatalf("open decryptor: %v", err)
	}
	if received != "gcp-kms://projects/harpy-fixture/locations/europe-west1/"+
		"keyRings/iam/cryptoKeys/totp/cryptoKeyVersions/7" {
		t.Fatalf("unexpected KMS key version %q", received)
	}
}

func TestOpenIAMTOTPDecryptorFailsClosed(t *testing.T) {
	if _, err := openIAMTOTPDecryptor(
		context.Background(), config.IAMAPI{}, nil,
	); err == nil {
		t.Fatal("invalid IAM KMS runtime was accepted")
	}
}
