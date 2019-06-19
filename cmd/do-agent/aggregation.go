package main

var dbaasAggregationSpec = map[string]string{

	"postgresql_pg_stat_user_tables_idx_scan":      "table_name",
	"postgresql_pg_stat_user_tables_n_tup_ins":     "table_name",
	"postgresql_pg_stat_user_tables_n_tup_upd":     "table_name",
	"postgresql_pg_stat_user_tables_seq_scan":      "table_name",
	"postgresql_pg_stat_user_tables_n_tup_del":     "table_name",
	"postgresql_pg_stat_user_tables_idx_tup_fetch": "table_name",
	"postgresql_pg_stat_user_tables_seq_tup_read":  "table_name",

	"mysql_perf_schema_table_io_waits_total_fetch":          "name",
	"mysql_perf_schema_table_io_waits_total_update":         "name",
	"mysql_perf_schema_table_io_waits_total_delete":         "name",
	"mysql_perf_schema_table_io_waits_total_insert":         "name",
	"mysql_perf_schema_table_io_waits_seconds_total_fetch":  "name",
	"mysql_perf_schema_table_io_waits_seconds_total_update": "name",
	"mysql_perf_schema_table_io_waits_seconds_total_delete": "name",
	"mysql_perf_schema_table_io_waits_seconds_total_insert": "name",
}
