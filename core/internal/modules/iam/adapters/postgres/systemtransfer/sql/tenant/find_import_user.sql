SELECT users.id,users.revision
FROM iam_user_identifiers AS identifier
JOIN iam_users AS users ON users.id=identifier.user_id
WHERE identifier.kind='email_v1'
  AND identifier.normalized_value=$1
  AND identifier.state<>'retired'
LIMIT 1
FOR KEY SHARE OF users
