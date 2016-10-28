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

package collector

import (
	"sort"
	"testing"

	"github.com/digitalocean/do-agent/metrics"
)

type stubMetricSet struct {
	Name string
	Opts []metrics.RegOpt
}

type stubRegistry struct {
	RegisterNameOpts []stubMetricSet
	RegisterResult   metrics.MetricRef
	AddCollectorFunc metrics.Collector
}

func (s *stubRegistry) Register(name string, opts ...metrics.RegOpt) metrics.MetricRef {
	set := stubMetricSet{
		Name: name,
		Opts: opts,
	}
	s.RegisterNameOpts = append(s.RegisterNameOpts, set)
	return s.RegisterResult
}

func (s *stubRegistry) AddCollector(f metrics.Collector) {
	s.AddCollectorFunc = f
}

func (s *stubRegistry) Report(r metrics.Reporter) {
	s.AddCollectorFunc(r)
}

type stubUpdateSet struct {
	Ref         metrics.MetricRef
	Value       float64
	LabelValues []string
}

type stubReporter struct {
	UpdateSet []stubUpdateSet
}

func (s *stubReporter) Update(ref metrics.MetricRef, value float64, labelValues ...string) {
	set := stubUpdateSet{
		Ref:         ref,
		Value:       value,
		LabelValues: labelValues,
	}
	s.UpdateSet = append(s.UpdateSet, set)
}

// Verify that the stubRegistry implements the metrics.Registry interface.
var _ metrics.Registry = (*stubRegistry)(nil)

// Verify that the stubReporter implements the metrics.Reporter interface.
var _ metrics.Reporter = (*stubReporter)(nil)

// compareStringsUnordered tests if two slices contain the same elements (in any order).
// If they do not contain the same elements, elements which were exclusively found in a and b will
// be returned.
//
// Example:
// ok, aExtra, bExtra := CompareStringsUnordered([]string{"hello", "good", "world"}, []string{"good", "bye"})
// will return:
//   ok -> false
//   aExtra -> []string{"hello", "world"}
//   bExtra -> []string{"bye"}
func compareStringsUnordered(a, b []string) (bool, []string, []string) {
	aSorted := make([]string, len(a))
	copy(aSorted, a)
	sort.Strings(aSorted)
	bSorted := make([]string, len(b))
	copy(bSorted, b)
	sort.Strings(bSorted)

	i := 0
	j := 0

	aExtra := []string{}
	bExtra := []string{}

	for {
		if i == len(aSorted) {
			for ; j < len(bSorted); j++ {
				bExtra = append(bExtra, bSorted[j])
			}
			break
		}
		if j == len(bSorted) {
			for ; i < len(aSorted); i++ {
				aExtra = append(aExtra, aSorted[i])
			}
			break
		}
		if aSorted[i] == bSorted[j] {
			i++
			j++
		} else if aSorted[i] < bSorted[j] {
			aExtra = append(aExtra, aSorted[i])
			i++
		} else {
			bExtra = append(bExtra, bSorted[j])
			j++
		}
	}
	ok := len(aExtra) == 0 && len(bExtra) == 0
	return ok, aExtra, bExtra
}

func testForMetricNames(t *testing.T, expectedNames, actualNames []string) {
	ok, namesNotFound, namesNotExpected := compareStringsUnordered(expectedNames, actualNames)

	if !ok && len(namesNotFound) > 0 {
		for i := range namesNotFound {
			t.Errorf("expected metric name not found: %s", namesNotFound[i])
		}
	}

	if !ok && len(namesNotExpected) > 0 {
		for i := range namesNotExpected {
			t.Errorf("unexpected metric name encountered: %s", namesNotExpected[i])
		}
	}
}
