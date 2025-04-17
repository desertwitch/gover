package configuration

import (
	"github.com/desertwitch/gover/internal/queue"
	"github.com/desertwitch/gover/internal/schema"
)

// PipelineConfiguration is a structure holding the operational pipelines.
type PipelineConfiguration struct {
	EnumerationPipelines map[string]schema.Pipeline[*queue.EnumerationTask] // map[sourceName]schema.Pipeline
	EvaluationPipelines  map[string]schema.Pipeline[*schema.Moveable]       // map[shareName]schema.Pipeline
	IOPipelines          map[string]schema.Pipeline[*schema.Moveable]       // map[targetName]schema.Pipeline
}

// AppConfiguration is the principal structure holding the application configuration.
type AppConfiguration struct {
	Pipelines *PipelineConfiguration
}

// NewAppConfiguration returns a pointer to a new [AppConfiguration].
func NewAppConfiguration() *AppConfiguration {
	return &AppConfiguration{
		Pipelines: &PipelineConfiguration{
			EnumerationPipelines: make(map[string]schema.Pipeline[*queue.EnumerationTask]),
			EvaluationPipelines:  make(map[string]schema.Pipeline[*schema.Moveable]),
			IOPipelines:          make(map[string]schema.Pipeline[*schema.Moveable]),
		},
	}
}
