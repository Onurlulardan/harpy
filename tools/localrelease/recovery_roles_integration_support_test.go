package main

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"
)

func assertRecoveryRoles(t *testing.T, ctx context.Context, dsn string) {
	t.Helper()
	connection, err := pgx.Connect(ctx, dsn)
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close(context.Background())
	rows, err := connection.Query(ctx, `SELECT rolname,rolcanlogin,
COALESCE(rolpassword LIKE 'SCRAM-SHA-256$%',false)
FROM pg_catalog.pg_authid
WHERE rolname=ANY($1::text[]) ORDER BY rolname`,
		[]string{"harpy_backup", "harpy_backup_reader", "harpy_restore",
			"harpy_reactivation_observer"})
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	expected := map[string][2]bool{
		"harpy_backup":                {true, true},
		"harpy_backup_reader":         {false, false},
		"harpy_restore":               {true, true},
		"harpy_reactivation_observer": {true, true},
	}
	count := 0
	for rows.Next() {
		var name string
		var login, scram bool
		if rows.Scan(&name, &login, &scram) != nil ||
			expected[name] != [2]bool{login, scram} {
			t.Fatalf("recovery role %s state = login:%t scram:%t", name, login, scram)
		}
		count++
	}
	if rows.Err() != nil || count != len(expected) {
		t.Fatalf("recovery role count = %d, want %d", count, len(expected))
	}
	var setOption bool
	err = connection.QueryRow(ctx, `SELECT membership.set_option
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
JOIN pg_catalog.pg_roles AS target ON target.oid=membership.roleid
WHERE member.rolname='harpy_backup'
AND target.rolname='harpy_backup_reader'`).Scan(&setOption)
	if err != nil || !setOption {
		t.Fatalf("backup reader SET membership is missing: %v", err)
	}
	err = connection.QueryRow(ctx, `SELECT membership.set_option
FROM pg_catalog.pg_auth_members AS membership
JOIN pg_catalog.pg_roles AS member ON member.oid=membership.member
JOIN pg_catalog.pg_roles AS target ON target.oid=membership.roleid
WHERE member.rolname='harpy_restore'
AND target.rolname='harpy_owner'`).Scan(&setOption)
	if err != nil || !setOption {
		t.Fatalf("restore owner SET membership is missing: %v", err)
	}
}
