SELECT EXISTS (
    SELECT 1 FROM catalog_catalog_components
    WHERE component_id = $1::uuid
      AND component_kind = $2::text
      AND owner_generation = $3::bigint
      AND writer_state = 'v2'
      AND lifecycle_state = 'ACTIVE'
)
