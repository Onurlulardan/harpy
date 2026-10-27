//go:build integration

package postgres

const formsImportRetryPriorFunctions = `
CREATE FUNCTION platform.platform_async_forms_import_read_v1(uuid,uuid,uuid)
RETURNS TABLE(
 job_state text,cancellation_requested boolean,attempts integer,
 max_attempts integer,completed_at timestamptz
) LANGUAGE sql AS $function$
 SELECT 'queued'::text,false,0,8,NULL::timestamptz;
$function$;
CREATE FUNCTION platform.platform_async_forms_import_control_v1(
 uuid,uuid,uuid,boolean,timestamptz
) RETURNS TABLE(
 job_state text,cancellation_requested boolean,
 completed_at timestamptz,changed boolean
) LANGUAGE sql AS $function$
 SELECT 'queued'::text,false,NULL::timestamptz,false;
$function$;
CREATE FUNCTION platform.platform_async_forms_import_current_lease_v1(
 uuid,uuid,uuid,uuid,uuid,bigint
) RETURNS boolean LANGUAGE sql AS $function$ SELECT false; $function$;
CREATE FUNCTION platform.platform_async_forms_import_complete_v1(
 uuid,uuid,uuid,uuid,uuid,bigint,timestamptz
) RETURNS boolean LANGUAGE sql AS $function$ SELECT false; $function$;
`

const formsImportRetryLineageWrappers = `
CREATE FUNCTION public.read_forms_import_job(
 p_tenant_id uuid,p_import_id uuid,p_job_id uuid
) RETURNS TABLE(
 job_state text,cancellation_requested boolean,attempts integer,
 max_attempts integer,completed_at timestamptz
) LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $function$
 SELECT * FROM platform.platform_async_forms_import_read_v1(
  p_tenant_id,p_import_id,p_job_id
 );
$function$;
CREATE FUNCTION public.control_forms_import_job(
 p_tenant_id uuid,p_import_id uuid,p_job_id uuid,
 p_allow_cancel boolean,p_observed_at timestamptz
) RETURNS TABLE(
 job_state text,cancellation_requested boolean,
 completed_at timestamptz,changed boolean
) LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $function$
 SELECT * FROM platform.platform_async_forms_import_control_v1(
  p_tenant_id,p_import_id,p_job_id,p_allow_cancel,p_observed_at
 );
$function$;
CREATE FUNCTION public.lease_forms_import_job(
 p_tenant_id uuid,p_import_id uuid,p_job_id uuid,p_attempt_id uuid,
 p_lease_owner_id uuid,p_lease_generation bigint
) RETURNS boolean
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $function$
 SELECT platform.platform_async_forms_import_current_lease_v1(
  p_tenant_id,p_import_id,p_job_id,p_attempt_id,
  p_lease_owner_id,p_lease_generation
 );
$function$;
CREATE FUNCTION public.complete_forms_import_job(
 p_tenant_id uuid,p_import_id uuid,p_job_id uuid,p_attempt_id uuid,
 p_lease_owner_id uuid,p_lease_generation bigint,p_observed_at timestamptz
) RETURNS boolean
LANGUAGE sql SECURITY DEFINER SET search_path=pg_catalog AS $function$
 SELECT platform.platform_async_forms_import_complete_v1(
  p_tenant_id,p_import_id,p_job_id,p_attempt_id,
  p_lease_owner_id,p_lease_generation,p_observed_at
 );
$function$;
REVOKE ALL ON FUNCTION public.read_forms_import_job(uuid,uuid,uuid) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.control_forms_import_job(
 uuid,uuid,uuid,boolean,timestamptz
) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.lease_forms_import_job(
 uuid,uuid,uuid,uuid,uuid,bigint
) FROM PUBLIC;
REVOKE ALL ON FUNCTION public.complete_forms_import_job(
 uuid,uuid,uuid,uuid,uuid,bigint,timestamptz
) FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.read_forms_import_job(
 uuid,uuid,uuid
) TO harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION public.control_forms_import_job(
 uuid,uuid,uuid,boolean,timestamptz
) TO harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION public.lease_forms_import_job(
 uuid,uuid,uuid,uuid,uuid,bigint
) TO harpy_v2_runtime;
GRANT EXECUTE ON FUNCTION public.complete_forms_import_job(
 uuid,uuid,uuid,uuid,uuid,bigint,timestamptz
) TO harpy_v2_runtime;
`
