SELECT version_id,object_key,object_version,object_generation,
       ciphertext_sha256,ciphertext_bytes,key_id
FROM documents_recovery_inventory_v1($1,$2)
