module github.com/common-fate/observability

go 1.16

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/benbjohnson/clock v1.2.0 // indirect
	github.com/sethvargo/go-envconfig v0.4.0
	github.com/shirou/gopsutil v3.21.10+incompatible // indirect
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	go.opentelemetry.io/collector/model v0.40.0
	go.opentelemetry.io/contrib/instrumentation/host v0.23.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.23.0
	go.opentelemetry.io/contrib/propagators/b3 v1.2.0
	go.opentelemetry.io/contrib/propagators/ot v1.2.0
	go.opentelemetry.io/otel v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.23.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0
	go.opentelemetry.io/otel/metric v0.23.0
	go.opentelemetry.io/otel/sdk v1.2.0
	go.opentelemetry.io/otel/sdk/metric v0.23.0
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.42.0
)
