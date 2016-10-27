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

package log

import "testing"

func TestLogToLevel(t *testing.T) {
	tests := []struct {
		label string
		want  level
		err   error
	}{
		{
			label: errorLabel,
			want:  errorLevel,
			err:   nil,
		},
		{
			label: infoLabel,
			want:  infoLevel,
			err:   nil,
		},
		{
			label: debugLabel,
			want:  debugLevel,
			err:   nil,
		},
		{
			label: "foo",
			want:  0,
			err:   ErrUnrecognizedLogLevel,
		},
	}

	for _, test := range tests {
		l, err := toLevel(test.label)
		if test.err != nil && err != test.err {
			t.Errorf("want=%+v got=%+v", test.err, err)
			continue
		}
		if l != test.want {
			t.Errorf("want=%d got=%d", test.want, l)
		}
	}
}

func TestSetLevel(t *testing.T) {
	setLevel(errorLevel)
	if logLevel != errorLevel {
		t.Error("log level not expected value")
	}
}
