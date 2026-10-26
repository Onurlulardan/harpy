WITH requested AS (
    SELECT DISTINCT item.requested_id AS id
    FROM pg_catalog.unnest($1::uuid[]) AS item(requested_id)
), ranked AS (
    SELECT link.id, link.work_resource_id, link.target_owner,
           link.target_kind, link.target_id, link.target_version_id,
           link.target_reference_epoch, link.permission_revision,
           link.row_version,
           pg_catalog.row_number() OVER (
               PARTITION BY link.work_resource_id
               ORDER BY CASE link.target_kind
                   WHEN 'FORM' THEN 0 WHEN 'FORM_RECORD' THEN 1 ELSE 2 END,
                   link.target_kind, link.id
           ) AS link_rank,
           pg_catalog.count(*) OVER (
               PARTITION BY link.work_resource_id
           ) AS link_count
    FROM work_resource_links AS link
    JOIN requested ON requested.id = link.work_resource_id
    WHERE link.work_resource_kind = 'TASK'
      AND link.state = 'ACTIVE'
)
SELECT id, work_resource_id, target_owner, target_kind, target_id,
       target_version_id, target_reference_epoch, permission_revision,
       row_version, link_count
FROM ranked
WHERE link_rank <= 3
ORDER BY work_resource_id, link_rank
