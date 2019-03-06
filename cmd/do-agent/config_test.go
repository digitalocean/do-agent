package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDisableCollectorsAddsCorrectFlags(t *testing.T) {
	// some params are added by init funcs in os files so reset it to test
	additionalParams = []string{}
	disabledCollectors = map[string]interface{}{}

	items := []string{"hello", "world"}
	flags := make([]string, len(items))
	for i, item := range items {
		flags[i] = disableCollectorFlag(item)
	}

	disableCollectors(items...)
	assert.EqualValues(t, flags, additionalParams)
}

func TestDisableCollectorsIsIdempotent(t *testing.T) {
	// some params are added by init funcs in os files so reset it to test
	additionalParams = []string{}
	disabledCollectors = map[string]interface{}{}

	items := []string{"hello", "world", "world"}
	flags := []string{
		disableCollectorFlag("hello"),
		disableCollectorFlag("world"),
	}

	disableCollectors(items...)
	assert.EqualValues(t, flags, additionalParams)
}

func TestParseKubuernetesClusterUUID(t *testing.T) {
	userData := `k8saas_role: kubelet
k8saas_master_domain_name: "ok.bou.ke"
k8saas_bootstrap_token: "123"
k8saas_proxy_token: "456"
k8saas_ca_cert: "CERT CERT CERT"
k8saas_etcd_ca: "CERT2 CERT2 CERT2"
k8saas_etcd_key: "ok\nwhatever"
k8saas_etcd_cert: "NEAT"
k8saas_overlay_subnet: "GOOD"
k8saas_cluster_uuid: "11111111-2222-3333-4444-555555555555"
k8saas_dns_service_ip: "YES"`

	parsed, err := parseKubernetesClusterUUID(userData)
	require.NoError(t, err)
	assert.EqualValues(t, "11111111-2222-3333-4444-555555555555", parsed)
}

func TestParseKubuernetesClusterUUIDMissing(t *testing.T) {
	userData := `k8saas_role: kubelet
k8saas_master_domain_name: "ok.bou.ke"
k8saas_bootstrap_token: "123"
k8saas_proxy_token: "456"
k8saas_ca_cert: "CERT CERT CERT"
k8saas_etcd_ca: "CERT2 CERT2 CERT2"
k8saas_etcd_key: "ok\nwhatever"
k8saas_etcd_cert: "NEAT"
k8saas_overlay_subnet: "GOOD"
k8saas_dns_service_ip: "YES"`

	parsed, err := parseKubernetesClusterUUID(userData)
	require.Error(t, err)
	require.Equal(t, err, errClusterUUIDNotFound)
	assert.EqualValues(t, "", parsed)
}
