package cursor

import (
	"bytes"
	"testing"
	"time"
)

func TestAuthenticatedGlobalSnapshotAnchor(t *testing.T) {
	fixture := baseContextFixture(t)
	reference, err := buildContext(t, fixture).Catalog()
	if err != nil {
		t.Fatal(err)
	}
	context, err := NewGlobalContext(
		fixture.tenant, "", "public", "billing.catalog.list", testNow,
		fixture.auth, reference, fixture.query, fixture.mandatory,
		fixture.pageSize, fixture.sorts,
	)
	if err != nil {
		t.Fatal(err)
	}
	codec, err := newCodec(
		testKeyring(t, "current", 1, "previous", 2),
		bytes.NewReader(bytes.Repeat([]byte{3}, nonceBytes)),
		func() time.Time { return testNow.Add(100 * time.Millisecond) },
	)
	if err != nil {
		t.Fatal(err)
	}
	token, err := codec.Encode(context, []LastValue{scalarLast(t, 1, "record-a")})
	if err != nil {
		t.Fatal(err)
	}
	anchor, err := codec.AuthenticatedGlobalSnapshotAt(token)
	if err != nil || !anchor.Equal(testNow) {
		t.Fatalf("anchor = %v, %v", anchor, err)
	}
	segments := []byte(token)
	segments[len(segments)-1] ^= 1
	if _, err := codec.AuthenticatedGlobalSnapshotAt(string(segments)); err == nil {
		t.Fatal("tampered global cursor yielded a snapshot anchor")
	}
	wrongEnvironment, _ := NewGlobalContext(
		"99999999-9999-4999-8999-999999999999", "", "public",
		"billing.catalog.list", testNow, fixture.auth, reference, fixture.query,
		fixture.mandatory, fixture.pageSize, fixture.sorts,
	)
	wrongOperation, _ := NewGlobalContext(
		fixture.tenant, "", "public", "billing.catalog.get", testNow,
		fixture.auth, reference, fixture.query, fixture.mandatory,
		fixture.pageSize, fixture.sorts,
	)
	authenticated, _ := NewGlobalContext(
		fixture.tenant, fixture.subject, "authenticated",
		"billing.catalog.list", testNow, fixture.auth, reference, fixture.query,
		fixture.mandatory, fixture.pageSize, fixture.sorts,
	)
	for _, wrong := range []Context{wrongEnvironment, wrongOperation, authenticated} {
		if _, err := codec.Decode(token, wrong); err == nil {
			t.Fatal("global cursor identity transplant was accepted")
		}
	}
}

func TestTenantCursorCannotSupplyGlobalSnapshotAnchor(t *testing.T) {
	codec, _, token := encodedFixture(t)
	if _, err := codec.AuthenticatedGlobalSnapshotAt(token); err == nil {
		t.Fatal("tenant cursor yielded a global snapshot anchor")
	}
	now, err := codec.SnapshotNow()
	if err != nil || !now.Equal(testNow) {
		t.Fatalf("snapshot now = %v, %v", now, err)
	}
}
