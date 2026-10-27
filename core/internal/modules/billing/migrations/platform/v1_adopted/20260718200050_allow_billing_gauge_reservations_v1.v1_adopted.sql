-- +goose Up
ALTER TABLE platform.billing_usage_totals
    DROP CONSTRAINT billing_usage_totals_value_check;

ALTER TABLE platform.billing_usage_totals
    ADD CONSTRAINT billing_usage_totals_value_check CHECK (
        committed_total >= 0 AND reserved_total >= 0 AND gauge_total >= 0
        AND revision > 0
        AND ((meter_kind = 'counter' AND gauge_total = 0)
            OR (meter_kind = 'gauge' AND committed_total = 0))
    ) NOT VALID;

ALTER TABLE platform.billing_usage_totals
    VALIDATE CONSTRAINT billing_usage_totals_value_check;
