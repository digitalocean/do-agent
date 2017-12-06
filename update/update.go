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
	"os"
	"runtime"
	"strings"
	"time"

	"github.com/digitalocean/do-agent/config"

	"github.com/flynn/go-tuf/client"
	tufdata "github.com/flynn/go-tuf/data"
)

const (
	rootKey = `[{"keytype":"ed25519","keyval":{"public":"288a52b43e69e1c1a1a5552707aa50900e59cfead3870d2d17f9a49a48d9407c"}}]`

	// RepoURL is the remote TUF repository URL
	RepoURL = "https://repos.sonar.digitalocean.com/tuf"

	// Interval is the time in seconds between update checks
	Interval = 3600

	// RepoLocalStore is the local repository path
	RepoLocalStore = "/var/opt/digitalocean/do-agent"

	backupExt = ".bak"
)

// Updater is an interface that checks for updates of the running
// binary and exec's the new binary once updated. A boolean value is
// provided as a flag to force updates from development versions to
// whatever production version is available in the repositories.
type Updater interface {
	FetchLatestAndExec(bool) error
	FetchLatest(bool) error
	Interval() time.Duration
}

// Update manages the communication with the local repository, the
// remote tuf repository and the file manipulations that happen.
type update struct {
	localStorePath string
	repositoryURL  string
	interval       uint
	client         *client.Client
}

// NewUpdate returns an Update object which has the components for a tuf client.
func NewUpdate(localStorePath, repositoryURL string, interval uint) Updater {
	return &update{
		localStorePath: localStorePath,
		repositoryURL:  repositoryURL,
		interval:       interval,
	}
}

func (u *update) createTufClient() (*client.Client, error) {
	if u.client != nil {
		return u.client, nil
	}

	localStoreFile := fmt.Sprintf("%s%s", u.localStorePath, "/tufLocalStore")
	ls, err := client.FileLocalStore(localStoreFile)
	if err != nil {
		return nil, ErrUnableToCreateLocalStore{Path: localStoreFile}
	}

	rs, err := client.HTTPRemoteStore(u.repositoryURL, nil)
	if err != nil {
		return nil, ErrUnableToQueryRemoteStore{StoreURL: u.repositoryURL}
	}

	tc := client.NewClient(ls, rs)
	u.client = tc
	return tc, nil
}

func (u *update) downloadTarget(target string, tmp Destination) error {
	err := u.client.Download(target, tmp)
	if err != nil {
		return ErrDownloadingTarget{Reason: err.Error()}
	}
	return nil
}

func (u *update) prepareTufRepository(rootKeys []*tufdata.Key) error {
	err := u.client.Init(rootKeys, len(rootKeys))
	if err != nil {
		return ErrUnableToInitializeRepo{Reason: err.Error()}
	}

	if _, err = u.client.Update(); err != nil && !client.IsLatestSnapshot(err) {
		return ErrUnableToUpdateRepo{Reason: err.Error()}
	}
	return nil
}

func (u *update) findUpdates(forceUpdate bool) (string, error) {
	baseVersion := fmt.Sprintf("/do-agent/do-agent_%s_%s_", runtime.GOOS, runtime.GOARCH)

	targets, err := u.client.Targets()
	if err != nil {
		return "", ErrUnableToRetrieveTargets
	}

	for target := range targets {
		if strings.Contains(target, baseVersion) {
			newVersion := strings.TrimLeft(target, baseVersion)
			if upgradeVersion(config.Version(), newVersion, forceUpdate) {
				return target, nil
			}
		}
	}
	return "", ErrUpdateNotAvailable
}

// FetchLatestAndExec fetches the lastest do-agent binary, replaces
// the running binary and calls exec on the new binary.
func (u *update) FetchLatestAndExec(forceUpdate bool) error {
	// record bin path before update
	binPathOrig, err := currentExecPath()
	if err != nil {
		return ErrUnableToDetermineRunningProcess{Reason: err.Error()}
	}

	binBackupPath := fmt.Sprintf("%s%s", binPathOrig, backupExt)

	// delete any pre-existing backup files
	if _, err := os.Stat(binBackupPath); err == nil {
		_ = os.Remove(binBackupPath)
	}

	if err := u.FetchLatest(forceUpdate); err != nil {
		return err
	}

	binPath, err := currentExecPath()
	if err != nil {
		return ErrUnableToDetermineRunningProcess{Reason: err.Error()}
	}

	if err := executeBinary(binPathOrig); err == nil {
		return nil
	}

	// when downloaded binary fails to exec, replace it with the
	// backed up binary. Update the running binary path.
	if err := os.Rename(binBackupPath, binPath); err != nil {
		return errors.New("restoring backed up file")
	}
	return ErrExecuteBinary
}

// FetchLatest fetches the lastest do-agent binary and replace the
// running binary. In the event of an error it attempts to rollback to
// the previous version of do-agent.
func (u *update) FetchLatest(forceUpdate bool) error {
	if _, err := u.createTufClient(); err != nil {
		return err
	}

	tempFile, err := NewTempFile(u.localStorePath, "temp_tuf")
	if err != nil {
		return ErrUnableToCreateTempfile{Path: err.Error()}
	}
	defer tempFile.Delete()

	rootKeys, err := parseKeys(rootKey)
	if err != nil {
		return err
	}

	if err = u.prepareTufRepository(rootKeys); err != nil {
		return err
	}

	upgradeTarget, err := u.findUpdates(forceUpdate)
	if err != nil {
		return err
	}

	if err := u.downloadTarget(upgradeTarget, tempFile); err != nil {
		return err
	}

	curFilePath, err := currentExecPath()
	if err != nil {
		return ErrUnableToDetermineRunningProcess{Reason: err.Error()}
	}

	curFilePathBackup := curFilePath + backupExt

	// move current file to backup location
	if err := os.Rename(curFilePath, curFilePathBackup); err != nil {
		return err
	}

	// copy downloaded file to current binary location
	if err := copyFile(tempFile.Name(), curFilePath); err == nil {
		return nil
	}

	return nil
}

// Interval returns the update interval
func (u *update) Interval() time.Duration {
	return time.Duration(u.interval) * time.Second
}
