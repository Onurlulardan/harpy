-- +goose Up
-- Migration stream version: 20260727220001.

SELECT 1 / CASE WHEN EXISTS (
    SELECT 1
    FROM platform.domain_cutovers AS domain
    WHERE domain.scope_kind = 'tenant'
      AND domain.domain = 'analytics'
      AND (domain.state <> 'v2' OR domain.writer <> 'v2')
) THEN 0 ELSE 1 END;
