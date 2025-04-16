package processors

import "github.com/desertwitch/gover/internal/schema"

// Pipeline is the principal implementation of a [schema.Pipeline].
type Pipeline struct {
	preProcessors  []schema.BatchProcessor
	mProcessors    []schema.Processor
	postProcessors []schema.BatchProcessor
}

// Add takes a [schema.Processor] and adds it to the pipeline.
func (p *Pipeline) Add(pr schema.Processor) schema.Pipeline { //nolint:ireturn
	p.mProcessors = append(p.mProcessors, pr)

	return p
}

// AddPreProcess takes a [schema.BatchProcessor] pre-processor and adds it to
// the pipeline.
func (p *Pipeline) AddPreProcess(bp schema.BatchProcessor) schema.Pipeline { //nolint:ireturn
	p.preProcessors = append(p.preProcessors, bp)

	return p
}

// AddPostProcess takes a [schema.BatchProcessor] post-processor and adds it to
// the pipeline.
func (p *Pipeline) AddPostProcess(bp schema.BatchProcessor) schema.Pipeline { //nolint:ireturn
	p.postProcessors = append(p.postProcessors, bp)

	return p
}

// Process sequentially processes all previously added [schema.Processor].
func (p *Pipeline) Process(m *schema.Moveable) bool {
	for _, fn := range p.mProcessors {
		if ok := fn(m); !ok {
			return false
		}
	}

	return true
}

// PreProcess sequentially processes all previously added
// [schema.BatchProcessor] pre-processors.
func (p *Pipeline) PreProcess(moveables []*schema.Moveable) ([]*schema.Moveable, bool) {
	movbls := make([]*schema.Moveable, len(moveables))
	copy(movbls, moveables)

	for _, fn := range p.preProcessors {
		if ms, ok := fn(movbls); ok {
			movbls = make([]*schema.Moveable, len(ms))
			copy(movbls, ms)
		} else {
			return nil, false
		}
	}

	return movbls, true
}

// PostProcess sequentially processes all previously added
// [schema.BatchProcessor] post-processors.
func (p *Pipeline) PostProcess(moveables []*schema.Moveable) ([]*schema.Moveable, bool) {
	movbls := make([]*schema.Moveable, len(moveables))
	copy(movbls, moveables)

	for _, fn := range p.postProcessors {
		if ms, ok := fn(movbls); ok {
			movbls = make([]*schema.Moveable, len(ms))
			copy(movbls, ms)
		} else {
			return nil, false
		}
	}

	return movbls, true
}
