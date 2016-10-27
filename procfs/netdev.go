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

package procfs

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

const networkPath = "/proc/net/dev"

// Network contains the data exposed by the /proc/net/dev psudo-file
// system file.
type Network struct {
	Interface    string
	RXBytes      uint64
	RXPackets    uint64
	RXErrs       uint64
	RXDrop       uint64
	RXFifo       uint64
	RXFrame      uint64
	RXCompressed uint64
	RXMulticast  uint64
	TXBytes      uint64
	TXPackets    uint64
	TXErrs       uint64
	TXDrop       uint64
	TXFifo       uint64
	TXColls      uint64
	TXCarrier    uint64
	TXCompressed uint64
}

// Networker is a collection of network metrics exposed by the
// procfs.
type Networker interface {
	NewNetwork() ([]Network, error)
}

// NewNetwork collects data from the /proc/net/dev psuedo-file system
// file and converts it into a Network struct.
func NewNetwork() ([]Network, error) {
	f, err := os.Open(networkPath)
	if err != nil {
		err = fmt.Errorf("Unable to collect network metrics from %s - error: %s", networkPath, err)
		return []Network{}, err
	}
	defer f.Close()

	return readNetwork(f)
}

func readNetwork(f io.Reader) ([]Network, error) {
	scanner := bufio.NewScanner(f)

	var networks []Network

	//Ignore the first two lines
	scanner.Scan()
	scanner.Scan()
	for scanner.Scan() {
		line := scanner.Text()

		network, err := parseNetwork(line)
		if err != nil {
			return []Network{}, err
		}
		networks = append(networks, network)
	}
	return networks, nil
}

// parseNetwork parses a string and returns a Network if the string is
// in the expected format.
func parseNetwork(line string) (Network, error) {
	lineArray := strings.Fields(line)

	if len(lineArray) != 17 {
		return Network{}, errors.New("Field mismatch error while parsing: " + networkPath)
	}

	network := Network{}
	network.Interface = strings.TrimRight(lineArray[0], ":")
	network.RXBytes, _ = strconv.ParseUint(lineArray[1], 10, 64)
	network.RXPackets, _ = strconv.ParseUint(lineArray[2], 10, 64)
	network.RXErrs, _ = strconv.ParseUint(lineArray[3], 10, 64)
	network.RXDrop, _ = strconv.ParseUint(lineArray[4], 10, 64)
	network.RXFifo, _ = strconv.ParseUint(lineArray[5], 10, 64)
	network.RXFrame, _ = strconv.ParseUint(lineArray[6], 10, 64)
	network.RXCompressed, _ = strconv.ParseUint(lineArray[7], 10, 64)
	network.RXMulticast, _ = strconv.ParseUint(lineArray[8], 10, 64)
	network.TXBytes, _ = strconv.ParseUint(lineArray[9], 10, 64)
	network.TXPackets, _ = strconv.ParseUint(lineArray[10], 10, 64)
	network.TXErrs, _ = strconv.ParseUint(lineArray[11], 10, 64)
	network.TXDrop, _ = strconv.ParseUint(lineArray[12], 10, 64)
	network.TXFifo, _ = strconv.ParseUint(lineArray[13], 10, 64)
	network.TXColls, _ = strconv.ParseUint(lineArray[14], 10, 64)
	network.TXCarrier, _ = strconv.ParseUint(lineArray[15], 10, 64)
	network.TXCompressed, _ = strconv.ParseUint(lineArray[16], 10, 64)

	return network, nil
}
