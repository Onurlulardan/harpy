SELECT iam_group_generation, iam_generation, target_generation
FROM platform.iam_v1_actor_assertion_issue_v1($1, $2);
