package pipelines

import (
	"time"

	"go.opentelemetry.io/otel/sdk/resource"
)

type PipelineConfig struct {
	Endpoint        string
	Insecure        bool
	Headers         map[string]string
	Resource        *resource.Resource
	ReportingPeriod string
	BatchTimeout    time.Duration
	Propagators     []string
}

type PipelineSetupFunc func(PipelineConfig) (func() error, error)
