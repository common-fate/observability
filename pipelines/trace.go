package pipelines

import (
	"context"
	"fmt"

	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/contrib/propagators/ot"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/encoding/gzip"
)

func NewTracePipeline(ctx context.Context, c PipelineConfig) (func(context.Context) error, error) {
	spanExporter, err := newTraceExporter(ctx, c.Endpoint, c.Insecure, c.Headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create span exporter: %v", err)
	}

	bsp := trace.NewBatchSpanProcessor(spanExporter, trace.WithBatchTimeout(c.BatchTimeout))
	tp := trace.NewTracerProvider(
		trace.WithSampler(trace.AlwaysSample()),
		trace.WithSpanProcessor(bsp),
		trace.WithResource(c.Resource),
	)

	if err = configurePropagators(c); err != nil {
		return nil, err
	}

	otel.SetTracerProvider(tp)

	return func(ctx context.Context) error {
		_ = bsp.Shutdown(ctx)
		return spanExporter.Shutdown(ctx)
	}, nil
}

func newTraceExporter(ctx context.Context, endpoint string, insecure bool, headers map[string]string) (*otlptrace.Exporter, error) {
	secureOption := otlptracegrpc.WithTLSCredentials(credentials.NewClientTLSFromCert(nil, ""))
	if insecure {
		secureOption = otlptracegrpc.WithInsecure()
	}
	return otlptrace.New(
		ctx,
		otlptracegrpc.NewClient(
			secureOption,
			otlptracegrpc.WithEndpoint(endpoint),
			otlptracegrpc.WithHeaders(headers),
			otlptracegrpc.WithCompressor(gzip.Name),
		),
	)
}

// configurePropagators configures B3 propagation by default
func configurePropagators(c PipelineConfig) error {
	propagatorsMap := map[string]propagation.TextMapPropagator{
		"b3":           b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader)),
		"baggage":      propagation.Baggage{},
		"tracecontext": propagation.TraceContext{},
		"ottrace":      ot.OT{},
	}
	var props []propagation.TextMapPropagator
	for _, key := range c.Propagators {
		prop := propagatorsMap[key]
		if prop != nil {
			props = append(props, prop)
		}
	}
	if len(props) == 0 {
		return fmt.Errorf("invalid configuration: unsupported propagators. Supported options: b3,baggage,tracecontext,ottrace")
	}
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		props...,
	))
	return nil
}
