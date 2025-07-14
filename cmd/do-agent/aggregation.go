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

	"kafka_log_log_size_value":                                        {"partition"},
	"kafka_server_brokertopicmetrics_bytesinpersec_count":             {"topic"},
	"kafka_server_brokertopicmetrics_bytesoutpersec_count":            {"topic"},
	"kafka_server_brokertopicmetrics_messagesinpersec_count":          {"topic"},
	"kafka_server_brokertopicmetrics_replicationbytesinpersec_count":  {"topic"},
	"kafka_server_brokertopicmetrics_replicationbytesoutpersec_count": {"topic"},

	"opensearch_cluster_health_active_shards":         {"name"},
	"opensearch_cluster_health_relocating_shards":     {"name"},
	"opensearch_cluster_health_unassigned_shards":     {"name"},
	"opensearch_cluster_health_number_of_nodes":       {"name"},
	"opensearch_cluster_health_active_primary_shards": {"name"},

	"opensearch_clusterstats_indices_count":               {"cluster_name"},
	"opensearch_clusterstats_indices_store_size_in_bytes": {"cluster_name"},
	"opensearch_clusterstats_indices_docs_count":          {"cluster_name"},
	"opensearch_clusterstats_indices_docs_deleted":        {"cluster_name"},

	"opensearch_indices_search_scroll_total":           {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_scroll_time_in_millis":  {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_query_total":            {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_query_time_in_millis":   {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_indexing_index_total":          {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_indexing_index_time_in_millis": {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_merges_total":                  {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_merges_total_time_in_millis":   {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_refresh_total":                 {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_refresh_total_time_in_millis":  {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_fetch_total":            {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_fetch_time_in_millis":   {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_suggest_total":          {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_search_suggest_time_in_millis": {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_query_cache_cache_size":        {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_query_cache_hit_count":         {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_query_cache_miss_count":        {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_indices_query_cache_evictions":         {"node_host", "node_attribute_zone", "cluster_name"},

	"opensearch_jvm_mem_heap_used_in_bytes": {"node_host", "node_attribute_zone", "cluster_name"},
	"opensearch_jvm_threads_count":          {"node_host", "node_attribute_zone", "cluster_name"},

	"opensearch_http_total_opened": {"node_host", "node_attribute_zone", "cluster_name"},
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

var mongoAggregationSpec = map[string][]string{
	"mongoagent_data_usage_percentage": {"cluster_uuid"},
}

// amdAggregatedLabels contains all the labels we want to aggregate on for AMD GPU metrics.
// keep only gpu_id and hostname
var amdAggregatedLabels = []string{
	"card_model", "card_series", "card_vendor", "cluster_name", "container",
	"driver_version", "gpu_compute_partition_type", "gpu_memory_partition_type",
	"gpu_partition_id", "gpu_uuid", "job_id", "job_partition", "job_user",
	"namespace", "pod", "serial_number", "usergroup_id", "vbios_version", "workload_id",
}

var gpuAggregationSpec = map[string][]string{
	// GPU Utilization metrics
	"amd_gpu_prof_gui_util_percent":      amdAggregatedLabels,
	"amd_gpu_prof_valu_pipe_issue_util":  amdAggregatedLabels,
	"amd_gpu_prof_tensor_active_percent": amdAggregatedLabels,
	"amd_gpu_prof_occupancy_percent":     amdAggregatedLabels,
	"amd_gpu_prof_fetch_size":            amdAggregatedLabels,
	"amd_gpu_prof_write_size":            amdAggregatedLabels,

	// GPU VRAM usage metrics
	"amd_gpu_used_vram":  amdAggregatedLabels,
	"amd_gpu_total_vram": amdAggregatedLabels,
	"amd_gpu_free_vram":  amdAggregatedLabels,

	// XGMI Bandwidth metrics - all neighbors 0-7
	"amd_xgmi_neighbor_0_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_1_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_2_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_3_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_4_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_5_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_6_tx_throughput": amdAggregatedLabels,
	"amd_xgmi_neighbor_7_tx_throughput": amdAggregatedLabels,

	"amd_xgmi_neighbor_0_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_1_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_2_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_3_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_4_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_5_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_6_response_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_7_response_tx": amdAggregatedLabels,

	"amd_xgmi_neighbor_0_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_1_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_2_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_3_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_4_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_5_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_6_request_tx": amdAggregatedLabels,
	"amd_xgmi_neighbor_7_request_tx": amdAggregatedLabels,

	// PCIe bandwidth
	"amd_pcie_bandwidth": amdAggregatedLabels,

	"amd_gpu_ecc_uncorrect_total":                         amdAggregatedLabels,
	"amd_pcie_replay_count":                               amdAggregatedLabels,
	"amd_pcie_recovery_count":                             amdAggregatedLabels,
	"amd_pcie_replay_rollover_count":                      amdAggregatedLabels,
	"amd_pcie_max_speed":                                  amdAggregatedLabels,
	"amd_pcie_speed":                                      amdAggregatedLabels,
	"amd_gpu_prof_cpf_cpf_stat_stall":                     amdAggregatedLabels,
	"amd_gpu_clock":                                       amdAggregatedLabels,
	"amd_gpu_violation_proc_hot_residency_accumulated":    amdAggregatedLabels,
	"amd_gpu_violation_soc_thermal_residency_accumulated": amdAggregatedLabels,
	"amd_gpu_violation_ppt_residency_accumulated":         amdAggregatedLabels,
	"amd_gpu_violation_hbm_thermal_residency_accumulated": amdAggregatedLabels,
	"amd_gpu_violation_vr_thermal_tracking_accumulated":   amdAggregatedLabels,
	"amd_gpu_junction_temperature":                        amdAggregatedLabels,
	"amd_gpu_power_usage":                                 amdAggregatedLabels,
	"amd_gpu_memory_temperature":                          amdAggregatedLabels,
}
