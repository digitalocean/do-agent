#!/bin/sh

# Copyright 2016 DigitalOcean
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#   http:#www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or
# implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# If "config" is passed as the argument, then send back the metric definition(s).
case $1 in
   config)
        cat <<'EOM'
{
  "protocol": "do-agent:1",
  "definitions": {
    "test": {
      "type": 1,
      "labels": {
        "user": "foo"
      }
    }
  }
}
EOM
        exit 0;;
esac

# Otherwise send the metric value(s).
cat <<'EOM'
{
  "metrics": {
    "test": {
      "value": 42.0
    }
  }
}
EOM


