package pipelines

import (
	"context"
	"fmt"
	"time"

	hostMetrics "go.opentelemetry.io/contrib/instrumentation/host"
	runtimeMetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	metricglobal "go.opentelemetry.io/otel/metric/global"
	controller "go.opentelemetry.io/otel/sdk/metric/controller/basic"
	processor "go.opentelemetry.io/otel/sdk/metric/processor/basic"
	selector "go.opentelemetry.io/otel/sdk/metric/selector/simple"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func NewMetricsPipeline(ctx context.Context, c PipelineConfig) (func(context.Context) error, error) {
	metricExporter, err := newMetricsExporter(ctx, c.Endpoint, c.Insecure, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create metric exporter: %v", err)
	}

	period := controller.DefaultPeriod
	if c.ReportingPeriod != "" {
		period, err = time.ParseDuration(c.ReportingPeriod)
		if err != nil {
			return nil, fmt.Errorf("invalid metric reporting period: %v", err)
		}
		if period <= 0 {

			return nil, fmt.Errorf("invalid metric reporting period: %v", c.ReportingPeriod)
		}
	}
	pusher := controller.New(
		processor.New(
			selector.NewWithInexpensiveDistribution(),
			metricExporter,
		),
		controller.WithExporter(metricExporter),
		controller.WithResource(c.Resource),
		controller.WithCollectPeriod(period),
	)

	if err = pusher.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start controller: %v", err)
	}

	provider := pusher.MeterProvider()

	if err = runtimeMetrics.Start(runtimeMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start runtime metrics: %v", err)
	}

	if err = hostMetrics.Start(hostMetrics.WithMeterProvider(provider)); err != nil {
		return nil, fmt.Errorf("failed to start host metrics: %v", err)
	}

	metricglobal.SetMeterProvider(provider)
	return func(ctx context.Context) error {
		_ = pusher.Stop(ctx)
		return metricExporter.Shutdown(ctx)
	}, nil
}

func newMetricsExporter(ctx context.Context, endpoint string, insecure bool, headers map[string]string) (*otlpmetric.Exporter, error) {
	secureOption := otlpmetricgrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlpmetricgrpc.WithInsecure()
	}
	return otlpmetric.New(
		ctx,
		otlpmetricgrpc.NewClient(
			secureOption,
			otlpmetricgrpc.WithEndpoint(endpoint),
			otlpmetricgrpc.WithHeaders(headers),
			otlpmetricgrpc.WithCompressor(gzip.Name),
		),
	)
}
