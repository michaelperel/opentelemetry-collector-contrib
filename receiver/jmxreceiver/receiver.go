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

package jmxreceiver

import (
	"context"
	"fmt"
	"net"

	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confignet"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver/otlpreceiver"
	"go.uber.org/zap"

	"github.com/open-telemetry/opentelemetry-collector-contrib/receiver/jmxreceiver/subprocess"
)

var _ component.MetricsReceiver = (*jmxMetricReceiver)(nil)

type jmxMetricReceiver struct {
	logger       *zap.Logger
	config       *config
	subprocess   *subprocess.Subprocess
	params       component.ReceiverCreateParams
	otlpReceiver component.MetricsReceiver
	nextConsumer consumer.MetricsConsumer
}

func newJMXMetricReceiver(
	params component.ReceiverCreateParams,
	config *config,
	nextConsumer consumer.MetricsConsumer,
) *jmxMetricReceiver {
	return &jmxMetricReceiver{
		logger:       params.Logger,
		params:       params,
		config:       config,
		nextConsumer: nextConsumer,
	}
}

func (jmx *jmxMetricReceiver) Start(ctx context.Context, host component.Host) (err error) {
	jmx.logger.Debug("Starting JMX Receiver")

	jmx.otlpReceiver, err = jmx.buildOTLPReceiver()
	if err != nil {
		return err
	}

	javaConfig, err := jmx.buildJMXMetricGathererConfig()
	if err != nil {
		return err
	}
	subprocessConfig := subprocess.Config{
		ExecutablePath: "java",
		Args:           []string{"-Dorg.slf4j.simpleLogger.defaultLogLevel=debug", "-jar", jmx.config.JARPath, "-config", "-"},
		StdInContents:  javaConfig,
	}

	jmx.subprocess = subprocess.NewSubprocess(&subprocessConfig, jmx.logger)

	err = jmx.otlpReceiver.Start(ctx, host)
	if err != nil {
		return err
	}

	return jmx.subprocess.Start(context.Background())
}

func (jmx *jmxMetricReceiver) Shutdown(ctx context.Context) error {
	jmx.logger.Debug("Shutting down JMX Receiver")
	subprocessErr := jmx.subprocess.Shutdown(ctx)
	otlpErr := jmx.otlpReceiver.Shutdown(ctx)
	if subprocessErr != nil {
		return subprocessErr
	}
	return otlpErr
}

func (jmx *jmxMetricReceiver) buildOTLPReceiver() (component.MetricsReceiver, error) {
	endpoint := jmx.config.OTLPExporterConfig.Endpoint
	host, port, err := net.SplitHostPort(endpoint)
	if err != nil {
		return nil, fmt.Errorf("failed to parse OTLPExporterConfig.Endpoint %s: %w", jmx.config.OTLPExporterConfig.Endpoint, err)
	}
	if port == "0" {
		// We need to know the port OTLP receiver will use to specify w/ java properties and not
		// rely on gRPC server's connection.
		listener, err := net.Listen("tcp", endpoint)
		if err != nil {
			return nil, fmt.Errorf(
				"failed determining desired port from OTLPExporterConfig.Endpoint %s: %w", jmx.config.OTLPExporterConfig.Endpoint, err,
			)
		}
		defer listener.Close()
		addr := listener.Addr().(*net.TCPAddr)
		port = fmt.Sprintf("%d", addr.Port)
		endpoint = fmt.Sprintf("%s:%s", host, port)
		jmx.config.OTLPExporterConfig.Endpoint = endpoint
	}

	factory := otlpreceiver.NewFactory()
	config := factory.CreateDefaultConfig().(*otlpreceiver.Config)
	config.GRPC.NetAddr = confignet.NetAddr{Endpoint: endpoint, Transport: "tcp"}
	config.HTTP = nil

	return factory.CreateMetricsReceiver(context.Background(), jmx.params, config, jmx.nextConsumer)
}

func (jmx *jmxMetricReceiver) buildJMXMetricGathererConfig() (string, error) {
	javaConfig := fmt.Sprintf(`otel.jmx.service.url = %v
otel.jmx.interval.milliseconds = %v
`, jmx.config.ServiceURL, jmx.config.CollectionInterval.Milliseconds())

	if jmx.config.TargetSystem != "" {
		javaConfig += fmt.Sprintf("otel.jmx.target.system = %v\n", jmx.config.TargetSystem)
	} else if jmx.config.GroovyScript != "" {
		javaConfig += fmt.Sprintf("otel.jmx.groovy.script = %v\n", jmx.config.GroovyScript)
	}

	javaConfig += fmt.Sprintf(`otel.exporter = otlp
otel.exporter.otlp.endpoint = %v
otel.exporter.otlp.metric.timeout = %v
`, jmx.config.OTLPExporterConfig.Endpoint, jmx.config.OTLPExporterConfig.Timeout.Milliseconds())

	if jmx.config.Username != "" {
		javaConfig += fmt.Sprintf("otel.jmx.username = %v\n", jmx.config.Username)
	}

	if jmx.config.Password != "" {
		javaConfig += fmt.Sprintf("otel.jmx.password = %v\n", jmx.config.Password)
	}

	return javaConfig, nil
}
