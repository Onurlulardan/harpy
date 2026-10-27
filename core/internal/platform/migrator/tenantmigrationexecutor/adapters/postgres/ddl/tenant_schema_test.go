package ddl

import "testing"

func TestCreateTenantSchemaIncludesActivationRuntimeACL(t *testing.T) {
	statement, err := CreateTenantSchema("t_abcdefghijklmnopqrstuvwxyz")
	if err != nil {
		t.Fatal(err)
	}
	want := `CREATE SCHEMA "t_abcdefghijklmnopqrstuvwxyz" AUTHORIZATION harpy_owner; ` +
		`REVOKE CREATE ON SCHEMA "t_abcdefghijklmnopqrstuvwxyz" FROM PUBLIC, ` +
		`harpy_runtime, harpy_v2_runtime; GRANT USAGE ON SCHEMA ` +
		`"t_abcdefghijklmnopqrstuvwxyz" TO harpy_runtime`
	if statement != want {
		t.Fatalf("tenant schema DDL = %q, want %q", statement, want)
	}
}

func TestCreateTenantSchemaRejectsUnsafeIdentifier(t *testing.T) {
	if _, err := CreateTenantSchema(`tenant"; DROP SCHEMA platform; --`); err == nil {
		t.Fatal("unsafe tenant schema identifier was accepted")
	}
}
