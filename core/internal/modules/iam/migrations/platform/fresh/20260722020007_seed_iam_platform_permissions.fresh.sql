-- +goose Up
INSERT INTO platform.iam_platform_permissions (
    id, permission_key, match_kind, resource_id, action_id, deployment_mode,
    delegable, step_up_required, owner_only, state, revision,
    created_at, updated_at, retired_at
)
SELECT expected.id, expected.permission_key, 'exact', resource.id, action.id,
       expected.deployment_mode, expected.delegable, expected.step_up_required,
       expected.owner_only, 'active', 1, '2026-07-22T00:00:00Z',
       '2026-07-22T00:00:00Z', NULL
FROM (VALUES
('14020000-0000-4000-8000-000000000001'::uuid,'platform.instance.read','platform.instance','read','both',true,false,false),
('14020000-0000-4000-8000-000000000002','platform.operators.read','platform.operators','read','both',true,false,false),
('14020000-0000-4000-8000-000000000003','platform.operators.grant','platform.operators','grant','both',false,true,true),
('14020000-0000-4000-8000-000000000004','platform.operators.revoke','platform.operators','revoke','both',false,true,true),
('14020000-0000-4000-8000-000000000005','platform.owner.transfer','platform.owner','transfer','both',false,true,true),
('14020000-0000-4000-8000-000000000006','platform.tenants.read','platform.tenants','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000007','platform.tenants.create','platform.tenants','create','saas',true,true,false),
('14020000-0000-4000-8000-000000000008','platform.tenants.update','platform.tenants','update','saas',true,false,false),
('14020000-0000-4000-8000-000000000009','platform.tenants.suspend','platform.tenants','suspend','saas',true,false,false),
('14020000-0000-4000-8000-000000000010','platform.tenants.reactivate','platform.tenants','reactivate','saas',true,false,false),
('14020000-0000-4000-8000-000000000011','platform.tenants.offboard','platform.tenants','offboard','saas',true,true,false),
('14020000-0000-4000-8000-000000000012','platform.tenants.purge','platform.tenants','purge','saas',true,true,false),
('14020000-0000-4000-8000-000000000013','platform.tenants.admin_setup.reissue','platform.tenants.admin_setup','reissue','saas',true,false,false),
('14020000-0000-4000-8000-000000000014','platform.database.stats.read','platform.database.stats','read','both',true,false,false),
('14020000-0000-4000-8000-000000000015','platform.database.schema.read','platform.database.schema','read','both',true,false,false),
('14020000-0000-4000-8000-000000000016','platform.database.diagnostics.read','platform.database.diagnostics','read','both',true,false,false),
('14020000-0000-4000-8000-000000000017','platform.recovery.read','platform.recovery','read','onpremise',true,false,false),
('14020000-0000-4000-8000-000000000018','platform.recovery.enable','platform.recovery','enable','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000019','platform.recovery.rotate','platform.recovery','rotate','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000020','platform.recovery.disable','platform.recovery','disable','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000021','platform.recovery.anchor.read','platform.recovery.anchor','read','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000022','platform.recovery.anchor.rotate','platform.recovery.anchor','rotate','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000023','platform.recovery_sets.read','platform.recovery_sets','read','onpremise',true,true,false),
('14020000-0000-4000-8000-000000000024','platform.recovery_sets.create','platform.recovery_sets','create','onpremise',true,true,false),
('14020000-0000-4000-8000-000000000025','platform.recovery_sets.manifest_download','platform.recovery_sets','manifest_download','onpremise',true,true,false),
('14020000-0000-4000-8000-000000000026','platform.recovery_sets.delete','platform.recovery_sets','delete','onpremise',true,true,false),
('14020000-0000-4000-8000-000000000027','platform.recovery_sets.restore','platform.recovery_sets','restore','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000028','platform.restore_operations.plan_download','platform.restore_operations','plan_download','onpremise',false,true,true),
('14020000-0000-4000-8000-000000000029','platform.billing.catalog.read','platform.billing.catalog','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000030','platform.billing.catalog.manage','platform.billing.catalog','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000031','platform.billing.subscriptions.read','platform.billing.subscriptions','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000032','platform.billing.subscriptions.manage','platform.billing.subscriptions','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000033','platform.billing.events.read','platform.billing','events.read','saas',true,false,false),
('14020000-0000-4000-8000-000000000034','platform.billing.revenue.read','platform.billing','revenue.read','saas',true,false,false),
('14020000-0000-4000-8000-000000000035','platform.billing.coupons.manage','platform.billing.coupons','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000036','platform.billing.campaigns.manage','platform.billing.campaigns','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000037','platform.analytics.usage.read','platform.analytics.usage','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000038','platform.bug_reports.read','platform.bug_reports','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000039','platform.bug_reports.manage','platform.bug_reports','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000040','platform.contact_messages.read','platform.contact_messages','read','saas',true,false,false),
('14020000-0000-4000-8000-000000000041','platform.contact_messages.manage','platform.contact_messages','manage','saas',true,false,false),
('14020000-0000-4000-8000-000000000042','platform.operations.health.read','platform.operations.health','read','both',true,false,false),
('14020000-0000-4000-8000-000000000043','platform.operations.alerts.read','platform.operations.alerts','read','both',true,false,false),
('14020000-0000-4000-8000-000000000044','platform.operations.alerts.manage','platform.operations.alerts','manage','both',true,false,false),
('14020000-0000-4000-8000-000000000045','platform.operations.traces.read','platform.operations.traces','read','both',true,false,false),
('14020000-0000-4000-8000-000000000046','platform.operations.permission_cache.flush','platform.operations.permission_cache','flush','both',true,true,false),
('14020000-0000-4000-8000-000000000047','platform.license.read','platform.license','read','onpremise',true,false,false),
('14020000-0000-4000-8000-000000000048','platform.license.refresh','platform.license','refresh','onpremise',true,true,false),
('14020000-0000-4000-8000-000000000052','platform.presence.read','platform.presence','read','saas',true,false,false)
) AS expected(id, permission_key, resource_key, action_key, deployment_mode,
              delegable, step_up_required, owner_only)
JOIN platform.iam_platform_resources AS resource
  ON resource.resource_key = expected.resource_key AND resource.state = 'active'
JOIN platform.iam_platform_actions AS action
  ON action.action_key = expected.action_key AND action.state = 'active';

INSERT INTO platform.iam_platform_permissions (
    id, permission_key, match_kind, resource_id, action_id, deployment_mode,
    delegable, step_up_required, owner_only, state, revision,
    created_at, updated_at, retired_at
) VALUES (
    '1402ffff-ffff-4fff-8fff-ffffffffffff', 'environment_all',
    'environment_all', NULL, NULL, 'both', false, false, true,
    'active', 1, '2026-07-22T00:00:00Z', '2026-07-22T00:00:00Z', NULL
);

SELECT 1 / CASE WHEN (
    SELECT pg_catalog.count(*) FROM platform.iam_platform_permissions
    WHERE state = 'active'
) = 50 THEN 1 ELSE 0 END AS iam_platform_permission_catalog_assertion;
