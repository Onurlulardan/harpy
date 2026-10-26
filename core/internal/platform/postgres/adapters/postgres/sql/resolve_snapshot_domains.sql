SELECT result_domain, result_generation
FROM platform.iam_snapshot_domain_generations_v1($1,$2,$3)
ORDER BY result_domain;
