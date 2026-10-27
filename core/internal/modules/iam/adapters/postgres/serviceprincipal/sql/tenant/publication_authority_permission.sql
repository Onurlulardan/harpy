SELECT COALESCE((
    SELECT permission.id::text
    FROM iam_permissions AS permission
    JOIN iam_resources AS resource ON resource.id = permission.resource_id
    JOIN iam_actions AS action ON action.id = permission.action_id
    WHERE permission.match_kind = 'exact'
      AND permission.is_protected
      AND permission.state = 'active'
      AND resource.resource_key = 'iam.capability_grant'
      AND resource.owner_domain = 'iam'
      AND resource.resource_kind = 'capability_grant'
      AND resource.is_protected
      AND resource.state = 'active'
      AND action.action_key = 'manage'
      AND action.state = 'active'
), '');
