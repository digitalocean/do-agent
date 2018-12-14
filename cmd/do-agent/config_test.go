package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
