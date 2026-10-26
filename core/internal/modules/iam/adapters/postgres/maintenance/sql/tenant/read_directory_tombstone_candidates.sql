SELECT identifier.id, identifier.revision
FROM iam_user_identifiers AS identifier
WHERE identifier.state = 'retired'
  AND identifier.retired_at >= identifier.cooldown_until
ORDER BY identifier.id
