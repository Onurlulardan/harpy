SELECT platform.billing_resolve_current_tenant_authority_v1($1::uuid),
       platform.billing_resolve_current_global_authority_v1('billing'::text);
