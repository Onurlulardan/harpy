SELECT state.installation_id, state.package_id, state.component_id,
       state.release_id, state.source_run_id, payload.package_slug,
       payload.package_version, state.content_sha256, state.state,
       payload.document_bytes, state.record_ids, state.revision, state.updated_at
FROM catalog_demo_states AS state
JOIN catalog_demo_payloads AS payload
  ON payload.component_id = state.component_id
 AND payload.package_id = state.package_id
 AND payload.release_id = state.release_id
WHERE state.installation_id = $1::uuid
ORDER BY state.component_id
