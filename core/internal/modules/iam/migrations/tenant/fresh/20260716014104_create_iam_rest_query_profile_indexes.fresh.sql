-- +goose Up
CREATE INDEX iam_users_profile_values_gin_idx
    ON iam_users USING gin (profile_values jsonb_path_ops);

CREATE INDEX iam_roles_profile_values_gin_idx
    ON iam_roles USING gin (profile_values jsonb_path_ops);

CREATE INDEX iam_organizations_profile_values_gin_idx
    ON iam_organizations USING gin (profile_values jsonb_path_ops);
