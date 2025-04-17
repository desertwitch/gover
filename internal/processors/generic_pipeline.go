package processors

import "github.com/desertwitch/gover/internal/schema"

// GenericPipeline is the principal generic implementation of a
// [schema.Pipeline].
type GenericPipeline[T any] struct {
	batchPreProcessors  []schema.BatchProcessor[T]
	itemProcessors      []schema.Processor[T]
	batchPostProcessors []schema.BatchProcessor[T]
}

// Add takes a [schema.Processor] and adds it to the pipeline.
func (p *GenericPipeline[T]) Add(processor schema.Processor[T]) schema.Pipeline[T] {
	p.itemProcessors = append(p.itemProcessors, processor)

	return p
}

// AddPreProcess takes a [schema.BatchProcessor] pre-processor and adds it to
// the pipeline.
func (p *GenericPipeline[T]) AddPreProcess(processor schema.BatchProcessor[T]) schema.Pipeline[T] {
	p.batchPreProcessors = append(p.batchPreProcessors, processor)

	return p
}

// AddPostProcess takes a [schema.BatchProcessor] post-processor and adds it to
// the pipeline.
func (p *GenericPipeline[T]) AddPostProcess(processor schema.BatchProcessor[T]) schema.Pipeline[T] {
	p.batchPostProcessors = append(p.batchPostProcessors, processor)

	return p
}

// Process sequentially runs all previously added [schema.Processor].
func (p *GenericPipeline[T]) Process(item T) bool {
	for _, fn := range p.itemProcessors {
		if success := fn(item); !success {
			return false
		}
	}

	return true
}

// PreProcess sequentially runs all previously added [schema.BatchProcessor]
// pre-processors.
func (p *GenericPipeline[T]) PreProcess(items []T) ([]T, bool) {
	itemsCopy := make([]T, len(items))
	copy(itemsCopy, items)

	for _, fn := range p.batchPreProcessors {
		if returnItems, success := fn(itemsCopy); success {
			itemsCopy = make([]T, len(returnItems))
			copy(itemsCopy, returnItems)
		} else {
			return nil, false
		}
	}

	return itemsCopy, true
}

// PostProcess sequentially runs all previously added [schema.BatchProcessor]
// post-processors.
func (p *GenericPipeline[T]) PostProcess(items []T) ([]T, bool) {
	itemsCopy := make([]T, len(items))
	copy(itemsCopy, items)

	for _, fn := range p.batchPostProcessors {
		if returnItems, success := fn(itemsCopy); success {
			itemsCopy = make([]T, len(returnItems))
			copy(itemsCopy, returnItems)
		} else {
			return nil, false
		}
	}

	return itemsCopy, true
}
