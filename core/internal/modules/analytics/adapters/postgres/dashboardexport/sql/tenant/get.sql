SELECT * FROM analytics_dashboard_export_snapshot_v1(
    (pg_catalog.current_schemas(false))[2]::name,$1,$2,$3,$4
)
