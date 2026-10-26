SELECT blocker
FROM catalog_adoption_blockers_v1($1::uuid)
ORDER BY blocker
