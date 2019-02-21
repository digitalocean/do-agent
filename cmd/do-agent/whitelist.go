package main

var k8sWhitelist = map[string]bool{
	"kube_deployment_spec_replicas":               true,
	"kube_deployment_status_replicas_available":   true,
	"kube_deployment_status_replicas_unavailable": true,

	"kube_daemonset_status_desired_number_scheduled": true,
	"kube_daemonset_status_number_available":         true,
	"kube_daemonset_status_number_unavailable":       true,

	"kube_statefulset_replicas":              true,
	"kube_statefulset_status_replicas_ready": true,

	"kube_node_status_allocatable": true,
	"kube_node_status_capacity":    true,
}

var postgresWhitelist = map[string]bool{

	"postgresql_pg_stat_activity_conn_count":       true,
	"postgresql_pg_stat_database_blks_hit":         true,
	"postgresql_pg_stat_database_blks_read":        true,
	"postgresql_pg_stat_database_deadlocks":        true,
	"postgresql_pg_stat_replication_bytes_diff":    true,
	"postgresql_pg_stat_user_tables_idx_scan":      true,
	"postgresql_pg_stat_user_tables_n_tup_ins":     true,
	"postgresql_pg_stat_user_tables_n_tup_upd":     true,
	"postgresql_pg_stat_user_tables_seq_scan":      true,
	"postgresql_pg_stat_user_tables_n_tup_del":     true,
	"postgresql_pg_stat_user_tables_idx_tup_fetch": true,
	"postgresql_pg_stat_user_tables_seq_tup_read":  true,
}
