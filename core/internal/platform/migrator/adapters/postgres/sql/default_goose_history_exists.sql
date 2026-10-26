SELECT EXISTS (
    SELECT 1
    FROM pg_catalog.pg_class AS relation_row
    JOIN pg_catalog.pg_namespace AS namespace_row
      ON namespace_row.oid = relation_row.relnamespace
    WHERE relation_row.relname = 'goose_db_version'
      AND relation_row.relkind = 'r'
);
