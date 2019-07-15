package main

var dropletAggregationSpec = map[string][]string{
	"sonar_cpu": {"cpu"},
}

var dbaasAggregationSpec = map[string][]string{

	"postgresql_pg_stat_user_tables_idx_scan":      {"table_name"},
	"postgresql_pg_stat_user_tables_n_tup_ins":     {"table_name"},
	"postgresql_pg_stat_user_tables_n_tup_upd":     {"table_name"},
	"postgresql_pg_stat_user_tables_seq_scan":      {"table_name"},
	"postgresql_pg_stat_user_tables_n_tup_del":     {"table_name"},
	"postgresql_pg_stat_user_tables_idx_tup_fetch": {"table_name"},
	"postgresql_pg_stat_user_tables_seq_tup_read":  {"table_name"},

	"mysql_perf_schema_table_io_waits_total_fetch":          {"name"},
	"mysql_perf_schema_table_io_waits_total_update":         {"name"},
	"mysql_perf_schema_table_io_waits_total_delete":         {"name"},
	"mysql_perf_schema_table_io_waits_total_insert":         {"name"},
	"mysql_perf_schema_table_io_waits_seconds_total_fetch":  {"name"},
	"mysql_perf_schema_table_io_waits_seconds_total_update": {"name"},
	"mysql_perf_schema_table_io_waits_seconds_total_delete": {"name"},
	"mysql_perf_schema_table_io_waits_seconds_total_insert": {"name"},

	"mysql_threads_connected": {"validate_password_dictionary_file_last_parsed"},
	"mysql_threads_created":   {"validate_password_dictionary_file_last_parsed"},
	"mysql_threads_running":   {"validate_password_dictionary_file_last_parsed"},

	"mysql_handler_read_first":    {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_key":      {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_last":     {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_next":     {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_prev":     {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_rnd":      {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
	"mysql_handler_read_rnd_next": {"innodb_buffer_pool_load_status", "innodb_buffer_pool_dump_status"},
}

var k8sAggregationSpec = map[string][]string{

	"kube_deployment_spec_replicas":               {"deployment", "namespace"},
	"kube_deployment_status_replicas_available":   {"deployment", "namespace"},
	"kube_deployment_status_replicas_unavailable": {"deployment", "namespace"},

	"kube_daemonset_status_desired_number_scheduled": {"daemonset", "namespace"},
	"kube_daemonset_status_number_available":         {"daemonset", "namespace"},
	"kube_daemonset_status_number_unavailable":       {"daemonset", "namespace"},

	"kube_statefulset_replicas":              {"statefulset", "namespace"},
	"kube_statefulset_status_replicas_ready": {"statefulset", "namespace"},
}
