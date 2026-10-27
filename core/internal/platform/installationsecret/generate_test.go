package installationsecret

import (
	"bytes"
	"encoding/base64"
	"testing"
)

func TestTemporaryPasswordUsesTwentyFourRandomBytesOnce(t *testing.T) {
	raw := bytes.Repeat([]byte{0x5a}, rawPasswordBytes)
	password, err := generateTemporaryPassword(bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	var retained []byte
	if err = password.Consume(func(wire []byte) error {
		if len(wire) != wirePasswordBytes || !validPasswordWire(wire) {
			t.Fatalf("invalid password wire %q", wire)
		}
		decoded, decodeErr := base64.RawURLEncoding.DecodeString(string(wire))
		if decodeErr != nil || !bytes.Equal(decoded, raw) {
			t.Fatal("password does not encode the exact random input")
		}
		retained = wire
		return nil
	}); err != nil {
		t.Fatal(err)
	}
	if err = password.Consume(func([]byte) error { return nil }); err != ErrInvalid {
		t.Fatalf("second consume err=%v", err)
	}
	if !bytes.Equal(retained, make([]byte, wirePasswordBytes)) {
		t.Fatal("consumer wire was not wiped")
	}
}

func TestTemporaryPasswordRejectsBrokenRandomness(t *testing.T) {
	if _, err := generateTemporaryPassword(bytes.NewReader(nil)); err != ErrInterrupted {
		t.Fatalf("err=%v", err)
	}
	if _, err := generateTemporaryPassword(bytes.NewReader(make([]byte, rawPasswordBytes))); err != ErrInterrupted {
		t.Fatalf("zero password err=%v", err)
	}
}
