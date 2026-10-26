SELECT organization.id,organization.revision
FROM iam_organizations AS organization
WHERE organization.slug=$1
LIMIT 1
FOR KEY SHARE
