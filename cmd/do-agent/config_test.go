package main

import (
	"testing"

	dto "github.com/prometheus/client_model/go"
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

func TestConvertLabelPairs(t *testing.T) {

	sPtr := func(s string) *string { return &s }
	pairs := convertToLabelPairs([]string{"user_id:1234"})
	require.Equal(t, []*dto.LabelPair{{Name: sPtr("user_id"), Value: sPtr("1234")}}, pairs)

	pairs = convertToLabelPairs([]string{"user_id:1234", "dbaas_cluster_uuid:ruiheiuqhf"})
	require.Equal(t, []*dto.LabelPair{{Name: sPtr("user_id"), Value: sPtr("1234")}, {Name: sPtr("dbaas_cluster_uuid"), Value: sPtr("ruiheiuqhf")}}, pairs)

	pairs = convertToLabelPairs([]string{"user_id:12:34:56"})
	require.Equal(t, []*dto.LabelPair{{Name: sPtr("user_id"), Value: sPtr("12:34:56")}}, pairs)

	pairs = convertToLabelPairs([]string{})
	require.Empty(t, pairs)

	pairs = convertToLabelPairs(nil)
	require.Empty(t, pairs)
}
