//go:build integration

package postgres

import (
	"context"
	"testing"
)

const (
	schemaSubmissionV2GroupGeneration   int64 = 41
	schemaSubmissionV2BillingGeneration int64 = 42
	schemaSubmissionV2FormsGeneration   int64 = 43
)

func installSchemaSubmissionV2Authority(t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, tenantID, authorityID string) {
	t.Helper()
	_, err := environment.pool.Exec(ctx, `
CREATE TABLE platform.billing_tenant_authorities(
    authority_id uuid PRIMARY KEY,tenant_id uuid NOT NULL,scope_kind text NOT NULL,
    scope_id uuid NOT NULL,authority_group_id text NOT NULL,
    authority_group_generation bigint NOT NULL,billing_domain text NOT NULL,
    billing_generation bigint NOT NULL,forms_domain text NOT NULL,
    forms_generation bigint NOT NULL,created_at timestamptz NOT NULL
);
REVOKE ALL ON platform.billing_tenant_authorities FROM PUBLIC;
CREATE FUNCTION platform.billing_resolve_current_tenant_authority_receipt_v1(
    p_tenant_id uuid
)
RETURNS TABLE(authority_id uuid,authority_group_id text,
    authority_group_generation bigint,billing_generation bigint,
    forms_generation bigint)
LANGUAGE plpgsql SECURITY DEFINER SET search_path=pg_catalog
AS $function$
BEGIN
    IF SESSION_USER IS DISTINCT FROM 'harpy_v2_runtime'
       OR CURRENT_USER IS DISTINCT FROM 'harpy_owner' THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='V2 receipt role is invalid';
    END IF;
    RETURN QUERY SELECT authority.authority_id,authority.authority_group_id,
        authority.authority_group_generation,authority.billing_generation,
        authority.forms_generation
    FROM platform.billing_tenant_authorities authority
    JOIN platform.tenants tenant ON tenant.id=authority.tenant_id
    JOIN platform.cutover_groups group_row
      ON group_row.scope_kind=authority.scope_kind
     AND group_row.scope_id=authority.scope_id
     AND group_row.group_id=authority.authority_group_id
    JOIN platform.domain_cutovers billing_row
      ON billing_row.scope_kind=authority.scope_kind
     AND billing_row.scope_id=authority.scope_id
     AND billing_row.domain=authority.billing_domain
    JOIN platform.domain_cutovers forms_row
      ON forms_row.scope_kind=authority.scope_kind
     AND forms_row.scope_id=authority.scope_id
     AND forms_row.domain=authority.forms_domain
    WHERE authority.tenant_id=p_tenant_id
      AND authority.scope_kind='tenant' AND authority.scope_id=p_tenant_id
      AND authority.authority_group_id='billing_forms'
      AND authority.billing_domain='billing' AND authority.forms_domain='forms'
      AND tenant.deleted_at IS NULL AND tenant.status IN ('ACTIVE','TRIAL')
      AND group_row.state='v2' AND group_row.writer='v2'
      AND group_row.generation=authority.authority_group_generation
      AND billing_row.state='v2' AND billing_row.writer='v2'
      AND billing_row.generation=authority.billing_generation
      AND forms_row.state='v2' AND forms_row.writer='v2'
      AND forms_row.generation=authority.forms_generation
    FOR SHARE OF authority,tenant,group_row,billing_row,forms_row;
    IF NOT FOUND THEN
        RAISE EXCEPTION USING ERRCODE='42501',MESSAGE='V2 receipt is unavailable';
    END IF;
END;
$function$;
REVOKE ALL ON FUNCTION
platform.billing_resolve_current_tenant_authority_receipt_v1(uuid) FROM PUBLIC;`)
	if err != nil {
		t.Fatalf("install V2 authority storage: %v", err)
	}
	actorID := schemaExecutorIntegrationUUID(995)
	_, err = environment.pool.Exec(ctx, `INSERT INTO platform.domain_cutovers(
scope_kind,scope_id,domain,state,writer,generation,requested_by,reason,change_reference)
VALUES ('tenant',$1,'billing','v2','v2',$2,$4,'test','P6.2-v2'),
	       ('tenant',$1,'forms','v2','v2',$3,$4,'test','P6.2-v2')`, tenantID,
		schemaSubmissionV2BillingGeneration, schemaSubmissionV2FormsGeneration,
		actorID)
	if err != nil {
		t.Fatalf("seed V2 authority domains: %v", err)
	}
	_, err = environment.pool.Exec(ctx, `INSERT INTO platform.cutover_groups(
scope_kind,scope_id,group_id,state,writer,generation,requested_by,reason,change_reference)
VALUES ('tenant',$1,'billing_forms','v2','v2',$2,$3,'test','P6.2-v2')`,
		tenantID, schemaSubmissionV2GroupGeneration, actorID)
	if err != nil {
		t.Fatalf("seed V2 authority group: %v", err)
	}
	_, err = environment.pool.Exec(ctx, `INSERT INTO platform.cutover_group_members(
scope_kind,scope_id,group_id,ordinal,domain)
VALUES ('tenant',$1,'billing_forms',1,'billing'),
	       ('tenant',$1,'billing_forms',2,'forms')`, tenantID)
	if err != nil {
		t.Fatalf("seed V2 authority members: %v", err)
	}
	_, err = environment.pool.Exec(ctx, `INSERT INTO platform.billing_tenant_authorities(
authority_id,tenant_id,scope_kind,scope_id,authority_group_id,
authority_group_generation,billing_domain,billing_generation,
forms_domain,forms_generation,created_at)
VALUES ($5,$1,'tenant',$1,'billing_forms',$2,'billing',$3,'forms',$4,
        pg_catalog.clock_timestamp())`, tenantID,
		schemaSubmissionV2GroupGeneration, schemaSubmissionV2BillingGeneration,
		schemaSubmissionV2FormsGeneration, authorityID)
	if err != nil {
		t.Fatalf("seed V2 authority: %v", err)
	}
}

func rotateSchemaSubmissionV2Authority(t *testing.T, ctx context.Context,
	environment *schemaExecutorIntegrationEnvironment, tenantID, authorityID string) {
	t.Helper()
	command, err := environment.pool.Exec(ctx, `UPDATE
platform.billing_tenant_authorities SET authority_id=$2 WHERE tenant_id=$1`,
		tenantID, authorityID)
	if err != nil || command.RowsAffected() != 1 {
		t.Fatalf("rotate V2 authority: rows=%d err=%v", command.RowsAffected(), err)
	}
}
