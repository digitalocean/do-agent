package main

// Label names shared by several metrics in the aggregation tables.
const (
	labelTableName                  = "table_name"
	labelName                       = "name"
	labelEventName                  = "event_name"
	labelPasswordDictParsed         = "validate_password_dictionary_file_last_parsed"
	labelInnodbBufferPoolLoadStatus = "innodb_buffer_pool_load_status"
	labelInnodbBufferPoolDumpStatus = "innodb_buffer_pool_dump_status"
	labelTopic                      = "topic"
	labelClusterName                = "cluster_name"
	labelNodeHost                   = "node_host"
	labelNodeAttributeZone          = "node_attribute_zone"
	labelDeployment                 = "deployment"
	labelNamespace                  = "namespace"
	labelDaemonset                  = "daemonset"
	labelContainer                  = "container"
)

var dropletAggregationSpec = map[string][]string{
	"sonar_cpu": {"cpu"},
}

var dbaasAggregationSpec = map[string][]string{

	"postgresql_pg_stat_user_tables_idx_scan":      {labelTableName},
	"postgresql_pg_stat_user_tables_n_tup_ins":     {labelTableName},
	"postgresql_pg_stat_user_tables_n_tup_upd":     {labelTableName},
	"postgresql_pg_stat_user_tables_seq_scan":      {labelTableName},
	"postgresql_pg_stat_user_tables_n_tup_del":     {labelTableName},
	"postgresql_pg_stat_user_tables_idx_tup_fetch": {labelTableName},
	"postgresql_pg_stat_user_tables_seq_tup_read":  {labelTableName},

	"mysql_perf_schema_table_io_waits_total_fetch":          {labelName},
	"mysql_perf_schema_table_io_waits_total_update":         {labelName},
	"mysql_perf_schema_table_io_waits_total_delete":         {labelName},
	"mysql_perf_schema_table_io_waits_total_insert":         {labelName},
	"mysql_perf_schema_table_io_waits_seconds_total_fetch":  {labelName},
	"mysql_perf_schema_table_io_waits_seconds_total_update": {labelName},
	"mysql_perf_schema_table_io_waits_seconds_total_delete": {labelName},
	"mysql_perf_schema_table_io_waits_seconds_total_insert": {labelName},

	// Advanced MySQL / Percona: collapse high-cardinality labels
	"mysql_global_status_commands_total":          {"command"},
	"mysql_global_status_handlers_total":          {"handler"},
	"mysql_global_status_innodb_row_ops_total":    {"operation"},
	"mysql_global_status_buffer_pool_pages":       {"state"},
	"mysql_perf_schema_file_events_total":         {labelEventName},
	"mysql_perf_schema_file_events_seconds_total": {labelEventName},
	"mysql_perf_schema_file_events_bytes_total":   {labelEventName},

	"mysql_threads_connected": {labelPasswordDictParsed},
	"mysql_threads_created":   {labelPasswordDictParsed},
	"mysql_threads_running":   {labelPasswordDictParsed},

	"mysql_handler_read_first":    {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_key":      {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_last":     {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_next":     {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_prev":     {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_rnd":      {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},
	"mysql_handler_read_rnd_next": {labelInnodbBufferPoolLoadStatus, labelInnodbBufferPoolDumpStatus},

	"kafka_log_log_size_value":                                        {"partition"},
	"kafka_server_brokertopicmetrics_bytesinpersec_count":             {labelTopic},
	"kafka_server_brokertopicmetrics_bytesoutpersec_count":            {labelTopic},
	"kafka_server_brokertopicmetrics_messagesinpersec_count":          {labelTopic},
	"kafka_server_brokertopicmetrics_replicationbytesinpersec_count":  {labelTopic},
	"kafka_server_brokertopicmetrics_replicationbytesoutpersec_count": {labelTopic},

	"opensearch_cluster_health_active_shards":         {labelName},
	"opensearch_cluster_health_relocating_shards":     {labelName},
	"opensearch_cluster_health_unassigned_shards":     {labelName},
	"opensearch_cluster_health_number_of_nodes":       {labelName},
	"opensearch_cluster_health_active_primary_shards": {labelName},

	"opensearch_clusterstats_indices_count":               {labelClusterName},
	"opensearch_clusterstats_indices_store_size_in_bytes": {labelClusterName},
	"opensearch_clusterstats_indices_docs_count":          {labelClusterName},
	"opensearch_clusterstats_indices_docs_deleted":        {labelClusterName},

	"opensearch_indices_search_scroll_total":           {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_scroll_time_in_millis":  {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_query_total":            {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_query_time_in_millis":   {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_indexing_index_total":          {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_indexing_index_time_in_millis": {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_merges_total":                  {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_merges_total_time_in_millis":   {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_refresh_total":                 {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_refresh_total_time_in_millis":  {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_fetch_total":            {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_fetch_time_in_millis":   {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_suggest_total":          {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_search_suggest_time_in_millis": {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_query_cache_cache_size":        {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_query_cache_hit_count":         {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_query_cache_miss_count":        {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_indices_query_cache_evictions":         {labelNodeHost, labelNodeAttributeZone, labelClusterName},

	"opensearch_jvm_mem_heap_used_in_bytes": {labelNodeHost, labelNodeAttributeZone, labelClusterName},
	"opensearch_jvm_threads_count":          {labelNodeHost, labelNodeAttributeZone, labelClusterName},

	"opensearch_http_total_opened": {labelNodeHost, labelNodeAttributeZone, labelClusterName},
}

var k8sAggregationSpec = map[string][]string{

	"kube_deployment_spec_replicas":               {labelDeployment, labelNamespace},
	"kube_deployment_status_replicas_available":   {labelDeployment, labelNamespace},
	"kube_deployment_status_replicas_unavailable": {labelDeployment, labelNamespace},

	"kube_daemonset_status_desired_number_scheduled": {labelDaemonset, labelNamespace},
	"kube_daemonset_status_number_available":         {labelDaemonset, labelNamespace},
	"kube_daemonset_status_number_unavailable":       {labelDaemonset, labelNamespace},

	"kube_statefulset_replicas":              {"statefulset", labelNamespace},
	"kube_statefulset_status_replicas_ready": {"statefulset", labelNamespace},
}

var mongoAggregationSpec = map[string][]string{
	"mongoagent_data_usage_percentage": {"cluster_uuid"},
}

// amdAggregatedLabels contains all the labels we want to aggregate on for AMD GPU metrics.
// keep only gpu_id and hostname
var amdAggregatedLabels = []string{
	"card_model", "card_series", "card_vendor", labelClusterName, labelContainer,
	"driver_version", "gpu_compute_partition_type", "gpu_memory_partition_type",
	"gpu_partition_id", "gpu_uuid", "job_id", "job_partition", "job_user",
	labelNamespace, "pod", "serial_number", "usergroup_id", "vbios_version", "workload_id",
	"link_index",
}

// nvidiaAggregatedLabels contains all the labels we want to aggregate on for NVIDIA GPU metrics.
// keep only gpu and Hostname
var nvidiaAggregatedLabels = []string{
	"pci_bus_id", "UUID", "device", "modelName", "DCGM_FI_DRIVER_VERSION",
	labelContainer, labelNamespace, "pod",
}

var gpuAggregationSpec = map[string][]string{
	// NVIDIA GPU metrics
	"dcgm_fi_dev_gpu_util":                          nvidiaAggregatedLabels,
	"dcgm_fi_dev_fb_free":                           nvidiaAggregatedLabels,
	"dcgm_fi_dev_fb_used":                           nvidiaAggregatedLabels,
	"dcgm_fi_dev_nvlink_bandwidth_total":            nvidiaAggregatedLabels,
	"dcgm_fi_dev_nvlink_crc_flit_error_count_total": nvidiaAggregatedLabels,
	"dcgm_fi_dev_nvlink_crc_data_error_count_total": nvidiaAggregatedLabels,
	"dcgm_fi_dev_nvlink_replay_error_count_total":   nvidiaAggregatedLabels,
	"dcgm_fi_dev_nvlink_recovery_error_count_total": nvidiaAggregatedLabels,
	"dcgm_fi_prof_sm_occupancy":                     nvidiaAggregatedLabels,
	"dcgm_fi_prof_pipe_tensor_active":               nvidiaAggregatedLabels,
	"dcgm_fi_prof_pcie_rx_bytes":                    nvidiaAggregatedLabels,
	"dcgm_fi_prof_pcie_tx_bytes":                    nvidiaAggregatedLabels,
	"dcgm_fi_dev_gpu_temp":                          nvidiaAggregatedLabels,
	"dcgm_fi_dev_memory_temp":                       nvidiaAggregatedLabels,
	"dcgm_fi_dev_power_usage":                       nvidiaAggregatedLabels,
	"dcgm_fi_dev_power_violation":                   nvidiaAggregatedLabels,
	"dcgm_fi_dev_thermal_violation":                 nvidiaAggregatedLabels,
	"dcgm_fi_dev_ecc_dbe_agg_total":                 nvidiaAggregatedLabels,
	"dcgm_fi_prof_sm_active":                        nvidiaAggregatedLabels,
	"dcgm_fi_prof_pipe_fp16_active":                 nvidiaAggregatedLabels,
	"dcgm_fi_prof_pipe_fp32_active":                 nvidiaAggregatedLabels,
	"dcgm_fi_prof_pipe_fp64_active":                 nvidiaAggregatedLabels,
	"dcgm_fi_dev_retired_sbe":                       nvidiaAggregatedLabels,
	"dcgm_fi_dev_retired_dbe":                       nvidiaAggregatedLabels,
	"dcgm_fi_dev_xid_errors":                        nvidiaAggregatedLabels,
	"dcgm_fi_dev_ecc_sbe_vol_total":                 nvidiaAggregatedLabels,
	"dcgm_fi_dev_ecc_dbe_vol_total":                 nvidiaAggregatedLabels,
	"dcgm_fi_dev_ecc_sbe_agg_total":                 nvidiaAggregatedLabels,
	"dcgm_fi_dev_pcie_replay_counter":               nvidiaAggregatedLabels,
	"dcgm_fi_dev_sm_clock":                          nvidiaAggregatedLabels,

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
	"amd_gpu_xgmi_link_rx": amdAggregatedLabels,
	"amd_gpu_xgmi_link_tx": amdAggregatedLabels,

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

	"amd_gpu_ecc_uncorrect_total":                            amdAggregatedLabels,
	"amd_gpu_ecc_correct_total":                              amdAggregatedLabels,
	"amd_pcie_replay_count":                                  amdAggregatedLabels,
	"amd_pcie_recovery_count":                                amdAggregatedLabels,
	"amd_pcie_replay_rollover_count":                         amdAggregatedLabels,
	"amd_pcie_max_speed":                                     amdAggregatedLabels,
	"amd_pcie_speed":                                         amdAggregatedLabels,
	"amd_pcie_nack_received_count":                           amdAggregatedLabels,
	"amd_pcie_nack_sent_count":                               amdAggregatedLabels,
	"amd_gpu_prof_cpf_cpf_stat_stall":                        amdAggregatedLabels,
	"amd_gpu_clock":                                          amdAggregatedLabels,
	"amd_gpu_violation_proc_hot_residency_accumulated":       amdAggregatedLabels,
	"amd_gpu_violation_processor_hot_residency_accumulated":  amdAggregatedLabels,
	"amd_gpu_violation_soc_thermal_residency_accumulated":    amdAggregatedLabels,
	"amd_gpu_violation_ppt_residency_accumulated":            amdAggregatedLabels,
	"amd_gpu_violation_hbm_thermal_residency_accumulated":    amdAggregatedLabels,
	"amd_gpu_violation_vr_thermal_tracking_accumulated":      amdAggregatedLabels,
	"amd_gpu_violation_socket_thermal_residency_accumulated": amdAggregatedLabels,
	"amd_gpu_violation_vr_thermal_residency_accumulated":     amdAggregatedLabels,
	"amd_gpu_violation_current_accumulated_counter":          amdAggregatedLabels,
	"amd_gpu_junction_temperature":                           amdAggregatedLabels,
	"amd_gpu_power_usage":                                    amdAggregatedLabels,
	"amd_gpu_package_power":                                  amdAggregatedLabels,
	"amd_gpu_memory_temperature":                             amdAggregatedLabels,
	"amd_gpu_gfx_activity":                                   amdAggregatedLabels,
	"amd_gpu_prof_sm_active":                                 amdAggregatedLabels,
}

// DI metrics: drop high-cardinality labels we don't want to keep.
// NOTE: do NOT include "le" because bucket-style series need it.
var diLabelsToDrop = []string{
	labelContainer,
	"job",
	"otel_scope_name",
	"otel_scope_schema_url",
	"otel_scope_version",
}

// diAggregationSpec lists DI metric names and which labels should be DROPPED.
// Anything NOT in this list will keep labels as-is.
var diAggregationSpec = map[string][]string{
	// DI GPU metrics
	"gradient_infra_di_gpu_junction_temperature":       diLabelsToDrop,
	"gradient_infra_di_gpu_total_vram":                 diLabelsToDrop,
	"gradient_infra_di_gpu_used_vram":                  diLabelsToDrop,
	"gradient_infra_di_gpu_free_vram":                  diLabelsToDrop,
	"gradient_infra_di_gpu_occupancy_percent":          diLabelsToDrop,
	"gradient_infra_di_gpu_tensor_utilization_percent": diLabelsToDrop,
	"gradient_infra_di_gpu_memory_temperature":         diLabelsToDrop,

	// Objective counters
	"gradient_infra_di_inference_objective_request_error_total": diLabelsToDrop,
	"gradient_infra_di_inference_objective_request_total":       diLabelsToDrop,

	// vLLM non-bucket metrics
	"gradient_infra_di_vllm:generation_tokens_total": diLabelsToDrop,
	"gradient_infra_di_vllm:num_requests_running":    diLabelsToDrop,
	"gradient_infra_di_vllm:num_requests_waiting":    diLabelsToDrop,
	"gradient_infra_di_vllm:kv_cache_usage_perc":     diLabelsToDrop,

	// vLLM bucket-style series
	"gradient_infra_di_vllm:e2e_request_latency_seconds_bucket":           diLabelsToDrop,
	"gradient_infra_di_vllm:inter_token_latency_seconds_bucket":           diLabelsToDrop,
	"gradient_infra_di_vllm:request_time_per_output_token_seconds_bucket": diLabelsToDrop,
	"gradient_infra_di_vllm:time_to_first_token_seconds_bucket":           diLabelsToDrop,
}
