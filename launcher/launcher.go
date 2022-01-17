package launcher

import (
	"context"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/common-fate/observability/pipelines"
	"github.com/sethvargo/go-envconfig"
	semconv "go.opentelemetry.io/collector/model/semconv/v1.5.0"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.uber.org/zap"
)

type Option func(*Config)

const (
	DefaultSpanExporterEndpoint   = "ingest.commonfate.io:443"
	DefaultMetricExporterEndpoint = "ingest.commonfate.io:443"
)

type Config struct {
	SpanExporterEndpoint           string `env:"OTEL_EXPORTER_OTLP_SPAN_ENDPOINT,default=ingest.commonfate.io:443"`
	SpanExporterEndpointInsecure   bool   `env:"OTEL_EXPORTER_OTLP_SPAN_INSECURE,default=false"`
	ServiceName                    string
	ServiceVersion                 string
	Headers                        map[string]string `env:"OTEL_EXPORTER_OTLP_HEADERS"`
	MetricExporterEndpoint         string            `env:"OTEL_EXPORTER_OTLP_METRIC_ENDPOINT,default=ingest.commonfate.io:443"`
	MetricExporterEndpointInsecure bool              `env:"OTEL_EXPORTER_OTLP_METRIC_INSECURE,default=false"`
	MetricsEnabled                 bool              `env:"OTEL_METRICS_ENABLED,default=true"`
	LogLevel                       string            `env:"OTEL_LOG_LEVEL,default=info"`
	Propagators                    []string          `env:"OTEL_PROPAGATORS,default=b3"`
	MetricReportingPeriod          string            `env:"OTEL_EXPORTER_OTLP_METRIC_PERIOD,default=30s"`
	BatchTimeout                   time.Duration
	resourceAttributes             map[string]string
	Resource                       *resource.Resource
	logger                         zap.Logger
	errorHandler                   otel.ErrorHandler
	context                        context.Context
}

func validateConfiguration(c Config) error {
	if len(c.ServiceName) == 0 {
		serviceNameSet := false
		for _, kv := range c.Resource.Attributes() {
			if kv.Key == semconv.AttributeServiceName {
				if len(kv.Value.AsString()) > 0 {
					serviceNameSet = true
				}
				break
			}
		}
		if !serviceNameSet {
			return errors.New("invalid configuration: service name missing. Configure WithServiceName in code")
		}
	}

	return nil
}

// WithMetricExporterEndpoint configures the endpoint for sending metrics via OTLP
func WithMetricExporterEndpoint(url string) Option {
	return func(c *Config) {
		c.MetricExporterEndpoint = url
	}
}

// WithSpanExporterEndpoint configures the endpoint for sending traces via OTLP
func WithSpanExporterEndpoint(url string) Option {
	return func(c *Config) {
		c.SpanExporterEndpoint = url
	}
}

// WithServiceName configures a "service.name" resource label
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion configures a "service.version" resource label
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithHeaders configures OTLP/gRPC connection headers
func WithHeaders(headers map[string]string) Option {
	return func(c *Config) {
		if c.Headers == nil {
			c.Headers = make(map[string]string)
		}
		for k, v := range headers {
			c.Headers[k] = v
		}
	}
}

// WithLogLevel configures the logging level for OpenTelemetry
func WithLogLevel(loglevel string) Option {
	return func(c *Config) {
		c.LogLevel = loglevel
	}
}

// WithSpanExporterInsecure permits connecting to the
// trace endpoint without a certificate
func WithSpanExporterInsecure(insecure bool) Option {
	return func(c *Config) {
		c.SpanExporterEndpointInsecure = insecure
	}
}

// WithMetricExporterInsecure permits connecting to the
// metric endpoint without a certificate
func WithMetricExporterInsecure(insecure bool) Option {
	return func(c *Config) {
		c.MetricExporterEndpointInsecure = insecure
	}
}

// WithResourceAttributes configures attributes on the resource
func WithResourceAttributes(attributes map[string]string) Option {
	return func(c *Config) {
		c.resourceAttributes = attributes
	}
}

// WithPropagators configures propagators
func WithPropagators(propagators []string) Option {
	return func(c *Config) {
		c.Propagators = propagators
	}
}

// Configures a global error handler to be used throughout an OpenTelemetry instrumented project.
// See "go.opentelemetry.io/otel"
func WithErrorHandler(handler otel.ErrorHandler) Option {
	return func(c *Config) {
		c.errorHandler = handler
	}
}

// WithMetricReportingPeriod configures the metric reporting period,
// how often the controller collects and exports metric data.
func WithMetricReportingPeriod(p time.Duration) Option {
	return func(c *Config) {
		c.MetricReportingPeriod = fmt.Sprint(p)
	}
}

// WithMetricEnabled configures whether metrics should be enabled
func WithMetricsEnabled(enabled bool) Option {
	return func(c *Config) {
		c.MetricsEnabled = enabled
	}
}

// WithBatchTimeout sets the batch timeout for sending traces to the collector
// https://pkg.go.dev/go.opentelemetry.io/otel/sdk@v0.13.0/trace#BatchSpanProcessorOptions
func WithBatchTimeout(timeout time.Duration) Option {
	return func(c *Config) {
		c.BatchTimeout = timeout
	}
}

// WithContext configures whether a custom context should be used
// to initiate tracing. If not, context.Background() is used.
func WithContext(ctx context.Context) Option {
	return func(c *Config) {
		c.context = ctx
	}
}

type defaultHandler struct {
	logger zap.Logger
}

func (l *defaultHandler) Handle(err error) {
	l.logger.Sugar().Debugf("error: %v\n", err)
}

func newConfig(opts ...Option) Config {
	var c Config
	envError := envconfig.Process(context.Background(), &c)
	c.BatchTimeout = 5 * time.Second
	c.logger = *zap.L()
	c.context = context.Background()
	c.errorHandler = &defaultHandler{logger: c.logger}
	var defaultOpts []Option

	for _, opt := range append(defaultOpts, opts...) {
		opt(&c)
	}
	c.Resource = newResource(&c)

	if envError != nil {
		c.logger.Sugar().Fatal(envError)
	}

	return c
}

type Launcher struct {
	config        Config
	shutdownFuncs []func(context.Context) error
}

func newResource(c *Config) *resource.Resource {
	r := resource.Environment()

	hostnameSet := false
	for iter := r.Iter(); iter.Next(); {
		if iter.Attribute().Key == semconv.AttributeHostName && len(iter.Attribute().Value.Emit()) > 0 {
			hostnameSet = true
		}
	}

	attributes := []attribute.KeyValue{
		attribute.String(semconv.AttributeTelemetrySDKName, "cfobservability"),
		attribute.String(semconv.AttributeTelemetrySDKLanguage, "go"),
		attribute.String(semconv.AttributeTelemetrySDKVersion, version),
	}

	if len(c.ServiceName) > 0 {
		attributes = append(attributes, attribute.String(semconv.AttributeServiceName, c.ServiceName))
	}

	if len(c.ServiceVersion) > 0 {
		attributes = append(attributes, attribute.String(semconv.AttributeServiceVersion, c.ServiceVersion))
	}

	for key, value := range c.resourceAttributes {
		if len(value) > 0 {
			if key == semconv.AttributeHostName {
				hostnameSet = true
			}
			attributes = append(attributes, attribute.String(key, value))
		}
	}

	if !hostnameSet {
		hostname, err := os.Hostname()
		if err != nil {
			c.logger.Sugar().Debugf("unable to set host.name. Set OTEL_RESOURCE_ATTRIBUTES=\"host.name=<your_host_name>\" env var or configure WithResourceAttributes in code: %v", err)
		} else {
			attributes = append(attributes, attribute.String(semconv.AttributeHostName, hostname))
		}
	}

	attributes = append(r.Attributes(), attributes...)

	// These detectors can't actually fail, ignoring the error.
	r, _ = resource.New(
		c.context,
		resource.WithSchemaURL(semconv.SchemaURL),
		resource.WithAttributes(attributes...),
	)

	// Note: There are new detectors we may wish to take advantage
	// of, now available in the default SDK (e.g., WithProcess(),
	// WithOSType(), ...).
	return r
}

func setupTracing(c Config) (func(ctx context.Context) error, error) {
	if c.SpanExporterEndpoint == "" {
		c.logger.Debug("tracing is disabled by configuration: no endpoint set")
		return nil, nil
	}
	return pipelines.NewTracePipeline(c.context, pipelines.PipelineConfig{
		Endpoint:     c.SpanExporterEndpoint,
		Insecure:     c.SpanExporterEndpointInsecure,
		Headers:      c.Headers,
		Resource:     c.Resource,
		Propagators:  c.Propagators,
		BatchTimeout: c.BatchTimeout,
	})
}

type setupFunc func(Config) (func(ctx context.Context) error, error)

func setupMetrics(c Config) (func(context.Context) error, error) {
	if !c.MetricsEnabled {
		c.logger.Debug("metrics are disabled by configuration: no endpoint set")
		return nil, nil
	}
	return pipelines.NewMetricsPipeline(c.context, pipelines.PipelineConfig{
		Endpoint:        c.MetricExporterEndpoint,
		Insecure:        c.MetricExporterEndpointInsecure,
		Headers:         c.Headers,
		Resource:        c.Resource,
		ReportingPeriod: c.MetricReportingPeriod,
		BatchTimeout:    c.BatchTimeout,
	})
}

func ConfigureOpentelemetry(opts ...Option) Launcher {
	c := newConfig(opts...)

	if c.LogLevel == "debug" {
		c.logger.Debug("debug logging enabled", zap.Any("configuration", c))
	}

	if c.Headers == nil {
		c.Headers = map[string]string{}
	}

	err := validateConfiguration(c)
	if err != nil {
		c.logger.Sugar().Fatalf("configuration error: %v", err)
	}

	if c.errorHandler != nil {
		otel.SetErrorHandler(c.errorHandler)
	}

	ls := Launcher{
		config: c,
	}

	for _, setup := range []setupFunc{setupTracing, setupMetrics} {
		shutdown, err := setup(c)
		if err != nil {
			c.logger.Sugar().Fatalf("setup error: %v", err)
			continue
		}
		if shutdown != nil {
			ls.shutdownFuncs = append(ls.shutdownFuncs, shutdown)
		}
	}
	return ls
}

func (ls Launcher) Shutdown() {
	ls.ShutdownContext(context.Background())
}

func (ls Launcher) ShutdownContext(ctx context.Context) {
	for _, shutdown := range ls.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			ls.config.logger.Sugar().Fatalf("failed to stop exporter: %v", err)
		}
	}
}
