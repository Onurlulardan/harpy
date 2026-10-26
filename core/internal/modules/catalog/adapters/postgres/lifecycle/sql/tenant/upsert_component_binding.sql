INSERT INTO catalog_component_bindings (
    installation_id, component_id, owner_module, owner_generation,
    owner_contract, content_sha256, state
) VALUES ($1::uuid, $2::uuid, $3::text, $4::bigint, $5::text, $6::text, $7::text)
ON CONFLICT (installation_id, component_id) DO UPDATE SET
    owner_module = EXCLUDED.owner_module,
    owner_generation = EXCLUDED.owner_generation,
    owner_contract = EXCLUDED.owner_contract,
    content_sha256 = EXCLUDED.content_sha256,
    state = EXCLUDED.state
