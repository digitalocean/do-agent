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
}

var dbaasWhitelist = map[string]bool{

	"postgresql_pg_stat_activity_conn_count":       true,
	"postgresql_pg_stat_database_blks_hit":         true,
	"postgresql_pg_stat_database_blks_read":        true,
	"postgresql_pg_stat_database_deadlocks":        true,
	"postgresql_pg_stat_replication_bytes_diff":    true,
	"postgresql_pg_stat_user_tables_idx_scan":      true,
	"postgresql_pg_stat_user_tables_seq_scan":      true,
	"postgresql_pg_stat_user_tables_n_tup_ins":     true,
	"postgresql_pg_stat_user_tables_n_tup_upd":     true,
	"postgresql_pg_stat_user_tables_n_tup_del":     true,
	"postgresql_pg_stat_user_tables_idx_tup_fetch": true,
	"postgresql_pg_stat_user_tables_seq_tup_read":  true,
	"postgresql_pg_stat_database_xact_commit":      true,
	"postgresql_pg_stat_database_xact_rollback":    true,
	"postgresql_database_size_database_size ":      true,

	"mysql_threads_created":   true,
	"mysql_threads_connected": true,
	"mysql_threads_running":   true,

	"mysql_handler_read_key":      true,
	"mysql_handler_read_first":    true,
	"mysql_handler_read_next":     true,
	"mysql_handler_read_prev":     true,
	"mysql_handler_read_last":     true,
	"mysql_handler_read_rnd":      true,
	"mysql_handler_read_rnd_next": true,

	"mysql_com_select": true,
	"mysql_com_insert": true,
	"mysql_com_update": true,
	"mysql_com_delete": true,

	"mysql_perf_schema_table_io_waits_total_fetch":  true,
	"mysql_perf_schema_table_io_waits_total_insert": true,
	"mysql_perf_schema_table_io_waits_total_update": true,
	"mysql_perf_schema_table_io_waits_total_delete": true,

	"mysql_perf_schema_table_io_waits_seconds_total_fetch":  true,
	"mysql_perf_schema_table_io_waits_seconds_total_insert": true,
	"mysql_perf_schema_table_io_waits_seconds_total_update": true,
	"mysql_perf_schema_table_io_waits_seconds_total_delete": true,

	"mysql_innodb_buffer_pool_reads":         true,
	"mysql_innodb_buffer_pool_read_requests": true,
	"mysql_innodb_data_written":              true,

	"mysql_questions":                true,
	"mysql_slow_queries":             true,
	"mysql_global_connection_memory": true,

	"redis_total_connections_received": true,
	"redis_rejected_connections":       true,
	"redis_evicted_keys":               true,
	"redis_keyspace_hits":              true,
	"redis_keyspace_keys":              true,
	"redis_keyspace_misses":            true,
	"redis_instantaneous_ops_per_sec":  true,
	"redis_used_memory_rss":            true,
	"redis_used_memory":                true,
	"redis_connected_slaves":           true,
	"redis_clients":                    true,
	"redis_total_commands_processed":   true,
	"redis_uptime":                     true,
	"redis_replication_lag":            true,
	"redis_replication_offset":         true,

	"kafka_log_Log_Size_Value": true,

	"kafka_server_BrokerTopicMetrics_BytesInPerSec_Count":             true,
	"kafka_server_BrokerTopicMetrics_BytesOutPerSec_Count":            true,
	"kafka_server_BrokerTopicMetrics_MessagesInPerSec_Count":          true,
	"kafka_server_BrokerTopicMetrics_ReplicationBytesInPerSec_Count":  true,
	"kafka_server_BrokerTopicMetrics_ReplicationBytesOutPerSec_Count": true,
	"kafka_server_KafkaServer_linux_disk_read_bytes_Value":            true,
	"kafka_server_KafkaServer_linux_disk_write_bytes_Value":           true,
	"kafka_server_ReplicaManager_UnderReplicatedPartitions_Value":     true,

	"kafka_network_RequestMetrics_RequestsPerSec_Count":    true,
	"kafka_network_RequestChannel_ResponseQueueSize_Value": true,
	"kafka_network_RequestMetrics_TotalTimeMs_Count":       true,

	"kafka_controller_KafkaController_ActiveControllerCount_Value":          true,
	"kafka_controller_KafkaController_OfflinePartitionsCount_Value":         true,
	"kafka_controller_KafkaController_PreferredReplicaImbalanceCount_Value": true,
	"kafka_controller_ControllerStats_LeaderElectionRateAndTimeMs_Count":    true,
}
