// Copyright 2016 DigitalOcean
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
// implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package update

import (
	"errors"
	"fmt"
)

var (
	// ErrRootKeyParseFailed invalid root key format error
	ErrRootKeyParseFailed = errors.New("invalid root key format")

	// ErrReadFileFailed read file error
	ErrReadFileFailed = errors.New("unable to read file")

	// ErrWriteFileFailed write file error
	ErrWriteFileFailed = errors.New("unable to write to file")

	// ErrExecuteBinary inability to execute binary error
	ErrExecuteBinary = errors.New("unable to execute binary")

	// ErrBackupBinary binary backup error
	ErrBackupBinary = errors.New("unable to backup binary")

	// ErrInvalidVersionFormat version parseing error
	ErrInvalidVersionFormat = errors.New("unable to parse version")

	// ErrUnableToRetrieveTargets target list retrieval failure
	ErrUnableToRetrieveTargets = errors.New("unable to retrieve target file list")

	// ErrUpdateNotAvailable could not find any updates
	ErrUpdateNotAvailable = errors.New("no updates available")
)

// ErrUnableToCreateLocalStore creating repository error
type ErrUnableToCreateLocalStore struct {
	Path string
}

func (e ErrUnableToCreateLocalStore) Error() string {
	return fmt.Sprintf("Update: Unable to create repository file: %s", e.Path)
}

// ErrUnableToQueryRemoteStore querying repository error
type ErrUnableToQueryRemoteStore struct {
	StoreURL string
}

func (e ErrUnableToQueryRemoteStore) Error() string {
	return fmt.Sprintf("Update: Unable to query repository file: %s", e.StoreURL)
}

// ErrUnableToCreateTempfile creating temp file error
type ErrUnableToCreateTempfile struct {
	Path string
}

func (e ErrUnableToCreateTempfile) Error() string {
	return fmt.Sprintf("Update: Unable to create temp file: %s", e.Path)
}

// ErrUnableToInitializeRepo initializing repository error
type ErrUnableToInitializeRepo struct {
	Reason string
}

func (e ErrUnableToInitializeRepo) Error() string {
	return fmt.Sprintf("Update: Error initializing repository: %s", e.Reason)
}

//ErrUnableToUpdateRepo updating repository error
type ErrUnableToUpdateRepo struct {
	Reason string
}

func (e ErrUnableToUpdateRepo) Error() string {
	return fmt.Sprintf("Update: Error updating repository: %s", e.Reason)
}

// ErrDownloadingTarget downloading target error
type ErrDownloadingTarget struct {
	Reason string
}

func (e ErrDownloadingTarget) Error() string {
	return fmt.Sprintf("Update: Error downloading target: %s", e.Reason)
}

// ErrUnableToDetermineRunningProcess determining running process error
type ErrUnableToDetermineRunningProcess struct {
	Reason string
}

func (e ErrUnableToDetermineRunningProcess) Error() string {
	return fmt.Sprintf("Update: Error determining running process: %s", e.Reason)
}
