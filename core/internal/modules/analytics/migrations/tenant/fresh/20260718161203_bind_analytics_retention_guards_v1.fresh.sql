-- +goose Up
CREATE OR REPLACE TRIGGER analytics_query_dependencies_immutable
BEFORE UPDATE OR DELETE ON analytics_query_dependencies
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_retention_v1();
ALTER TABLE analytics_query_dependencies ENABLE ALWAYS TRIGGER analytics_query_dependencies_immutable;

CREATE OR REPLACE TRIGGER analytics_publication_approvals_immutable
BEFORE UPDATE OR DELETE ON analytics_publication_approvals
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_retention_v1();
ALTER TABLE analytics_publication_approvals ENABLE ALWAYS TRIGGER analytics_publication_approvals_immutable;

CREATE OR REPLACE TRIGGER analytics_execution_audits_immutable
BEFORE UPDATE OR DELETE ON analytics_execution_audits
FOR EACH ROW EXECUTE FUNCTION analytics_reject_mutation_retention_v1();
ALTER TABLE analytics_execution_audits ENABLE ALWAYS TRIGGER analytics_execution_audits_immutable;

CREATE OR REPLACE TRIGGER analytics_operations_immutable
BEFORE INSERT OR UPDATE OR DELETE ON analytics_operations
FOR EACH ROW EXECUTE FUNCTION analytics_guard_operation_retention_v1();
ALTER TABLE analytics_operations ENABLE ALWAYS TRIGGER analytics_operations_immutable;

CREATE OR REPLACE TRIGGER analytics_saved_query_versions_guard
BEFORE UPDATE OR DELETE ON analytics_saved_query_versions
FOR EACH ROW EXECUTE FUNCTION analytics_guard_query_version_retention_v1();
ALTER TABLE analytics_saved_query_versions ENABLE ALWAYS TRIGGER analytics_saved_query_versions_guard;

CREATE OR REPLACE TRIGGER analytics_dashboard_versions_guard
BEFORE UPDATE OR DELETE ON analytics_dashboard_versions
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_version_retention_v1();
ALTER TABLE analytics_dashboard_versions ENABLE ALWAYS TRIGGER analytics_dashboard_versions_guard;

CREATE OR REPLACE TRIGGER analytics_dashboard_widgets_guard
BEFORE INSERT OR UPDATE OR DELETE ON analytics_dashboard_widgets
FOR EACH ROW EXECUTE FUNCTION analytics_guard_dashboard_widget_retention_v1();
ALTER TABLE analytics_dashboard_widgets ENABLE ALWAYS TRIGGER analytics_dashboard_widgets_guard;
