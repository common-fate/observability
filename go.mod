module github.com/common-fate/observability

go 1.16

require (
	github.com/StackExchange/wmi v1.2.1 // indirect
	github.com/felixge/httpsnoop v1.0.2
	github.com/go-chi/chi/v5 v5.0.7
	github.com/golang/protobuf v1.5.2
	github.com/sethvargo/go-envconfig v0.4.0
	github.com/shirou/gopsutil v3.21.10+incompatible // indirect
	github.com/stretchr/testify v1.7.0
	github.com/tklauser/go-sysconf v0.3.9 // indirect
	go.opentelemetry.io/collector/model v0.40.0
	go.opentelemetry.io/contrib v0.23.0
	go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc v0.28.0
	go.opentelemetry.io/contrib/instrumentation/host v0.23.0
	go.opentelemetry.io/contrib/instrumentation/runtime v0.23.0
	go.opentelemetry.io/contrib/propagators/b3 v1.3.0
	go.opentelemetry.io/contrib/propagators/ot v1.2.0
	go.opentelemetry.io/otel v1.3.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric v0.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc v0.26.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.2.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc v1.2.0
	go.opentelemetry.io/otel/metric v0.26.0
	go.opentelemetry.io/otel/sdk v1.3.0
	go.opentelemetry.io/otel/sdk/metric v0.26.0
	go.opentelemetry.io/otel/trace v1.3.0
	go.uber.org/zap v1.19.1
	google.golang.org/grpc v1.42.0
)
