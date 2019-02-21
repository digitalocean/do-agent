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

	"postgresql_pg_stat_database_blks_hit":                 true,
	"postgresql_pg_stat_database_blks_read":                true,
	"net_tcp_activeopens":                                  true,
	"postgresql_pg_stat_activity_conn_count":               true,
	"postgresql_pg_stat_activity_oldest_conn_age":          true,
	"postgresql_pg_stat_activity_oldest_running_query_age": true,
	"net_packets_recv":                                     true,
	"net_packets_sent":                                     true,
	"postgresql_pg_stat_database_deadlocks":                true,
	"postgresql_pg_stat_replication_flush_diff":            true,
}
