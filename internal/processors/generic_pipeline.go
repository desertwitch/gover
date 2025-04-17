package processors

import (
	"github.com/desertwitch/gover/internal/schema"
)

// GenericPipeline is the principal implementation of a [schema.Pipeline]. It
// holds all [schema.Processor] and [schema.BatchProcessor] functions for a
// specific operation and provides helper functions to execute them in
// sequential order and ensuring data safety.
//
// During execution, the pipeline exits on the first failed processor. The
// pipeline itself is not context-aware, rather the passed in [schema.Processor]
// and [schema.BatchProcessor] can themselves capture a context and handle its
// cancellation by returning "false" where and when early exit from the overall
// pipeline is warranted and wanted.
type GenericPipeline[T any] struct {
	batchPreProcessors  []schema.BatchProcessor[T]
	itemProcessors      []schema.Processor[T]
	batchPostProcessors []schema.BatchProcessor[T]
}

// Add takes a [schema.Processor] and adds it to the processing pipeline for
// later execution.
func (p *GenericPipeline[T]) Add(processor schema.Processor[T]) schema.Pipeline[T] {
	p.itemProcessors = append(p.itemProcessors, processor)

	return p
}

// AddPreProcess takes a [schema.BatchProcessor] pre-processor and adds it to
// the pre-processing pipeline for later execution.
func (p *GenericPipeline[T]) AddPreProcess(processor schema.BatchProcessor[T]) schema.Pipeline[T] {
	p.batchPreProcessors = append(p.batchPreProcessors, processor)

	return p
}

// AddPostProcess takes a [schema.BatchProcessor] post-processor and adds it to
// the post-processing pipeline for later execution.
func (p *GenericPipeline[T]) AddPostProcess(processor schema.BatchProcessor[T]) schema.Pipeline[T] {
	p.batchPostProcessors = append(p.batchPostProcessors, processor)

	return p
}

// Process sequentially runs all previously added [schema.Processor] processors
// on the given [T]. It is a helper method that is typically only called from
// within the structure responsible for the given T, such as a queue or
// queue-related processing function.
//
// The first failed processor will cause the function to return with "false".
func (p *GenericPipeline[T]) Process(item T) bool {
	for _, fn := range p.itemProcessors {
		if success := fn(item); !success {
			return false
		}
	}

	return true
}

// PreProcess sequentially runs all previously added [schema.BatchProcessor]
// pre-processors. It ensures only copies of the given slice are provided to the
// respective [schema.BatchProcessor] functions and itself eventually returns a
// copy of the original slice with all the manipulations. It is a helper method
// that is typically only called from within the structure responsible for the
// given slice of [T], such as a queue.
//
// The first failed processor will cause the function to return with "false".
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
// post-processors. It ensures only copies of the given slice are provided to
// the respective [schema.BatchProcessor] functions and itself eventually
// returns a copy of the original slice with all the manipulations. It is a
// helper method that is typically only called from within the structure
// responsible for the given slice of [T], such as a queue.
//
// The first failed processor will cause the function to return with "false".
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
