package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRegisterFilesystemFlagsRegistersFSTypesFlag(t *testing.T) {
	// this is initialized with the _linux.go file if run on linux, but
	// this test should run on all operating systems so we initialize it
	registerFilesystemFlags()
	assert.NotEmpty(t, additionalParams)
	assert.Contains(t, additionalParams, ignoredFSTypes)
}

func TestRegisterFilesystemFlagsRegistersMountPointFlag(t *testing.T) {
	// this is initialized with the _linux.go file if run on linux, but
	// this test should run on all operating systems so we initialize it
	registerFilesystemFlags()
	assert.NotEmpty(t, additionalParams)
	assert.Contains(t, additionalParams, ignoredMountPoints)
}
