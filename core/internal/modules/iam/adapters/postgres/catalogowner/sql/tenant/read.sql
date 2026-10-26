SELECT component.component_id::text,component.contract,
       COALESCE(candidate.content_sha256,component.content_sha256),
       CASE WHEN candidate.component_id IS NULL
            THEN component.lifecycle_state ELSE 'STAGED' END
FROM iam_catalog_components AS component
LEFT JOIN iam_catalog_component_candidates AS candidate
  ON candidate.component_id=component.component_id
 AND candidate.owner_generation=component.owner_generation
WHERE component.component_id=$1::uuid
  AND component.owner_generation=$2::bigint
  AND component.writer_state='v2'
