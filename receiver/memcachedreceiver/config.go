// Copyright 2020, OpenTelemetry Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package memcachedreceiver

import (
	"time"

	"go.opentelemetry.io/collector/config/configmodels"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/receiver/receiverhelper"
)

type config struct {
	configmodels.ReceiverSettings            `mapstructure:",squash"`
	receiverhelper.ScraperControllerSettings `mapstructure:",squash"`
	confignet.TCPAddr                        `mapstructure:",squash"`

	// Timeout for the memcache stats request
	Timeout time.Duration `mapstructure:"timeout"`
}
