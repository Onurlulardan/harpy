SELECT manage_permission.id, review_permission.id
FROM iam_permissions AS manage_permission
JOIN iam_resources AS manage_resource
  ON manage_resource.id = manage_permission.resource_id
JOIN iam_actions AS manage_action
  ON manage_action.id = manage_permission.action_id
CROSS JOIN iam_permissions AS review_permission
JOIN iam_resources AS review_resource
  ON review_resource.id = review_permission.resource_id
JOIN iam_actions AS review_action
  ON review_action.id = review_permission.action_id
WHERE manage_permission.match_kind = 'exact'
  AND manage_permission.is_protected
  AND manage_permission.state = 'active'
  AND manage_resource.resource_key = 'iam.permission_requests'
  AND manage_resource.state = 'active'
  AND manage_action.action_key = 'manage'
  AND manage_action.state = 'active'
  AND review_permission.match_kind = 'exact'
  AND review_permission.is_protected
  AND review_permission.state = 'active'
  AND review_resource.resource_key = 'iam.permission_requests'
  AND review_resource.state = 'active'
  AND review_action.action_key = 'review'
  AND review_action.state = 'active'
