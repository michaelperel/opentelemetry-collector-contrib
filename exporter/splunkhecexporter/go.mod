module github.com/open-telemetry/opentelemetry-collector-contrib/exporter/splunkhecexporter

go 1.14

require (
	github.com/census-instrumentation/opencensus-proto v0.3.0
	github.com/gogo/protobuf v1.3.1
	github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk v0.0.0-00010101000000-000000000000
	github.com/stretchr/testify v1.6.1
	go.opentelemetry.io/collector v0.14.1-0.20201106183657-c6b8f28c60b5
	go.uber.org/zap v1.16.0
	google.golang.org/grpc/examples v0.0.0-20200728194956-1c32b02682df // indirect
	google.golang.org/protobuf v1.25.0

)

replace github.com/open-telemetry/opentelemetry-collector-contrib/internal/splunk => ../../internal/splunk
