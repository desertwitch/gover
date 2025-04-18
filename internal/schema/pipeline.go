package schema

// Processor is a function that processes [T] as part of a [Pipeline].
//
// It should return a success boolean, any output the user needs to be aware of
// should be made with [Slog] calls, as no error type can be returned through
// the processor itself.
//
// During execution the pipeline should exit on the first failed processor, so
// "false" should be used only where pipeline failure is intended. The pipeline
// itself should not be context-aware, rather the Processor can capture a
// context and handle itself its cancellation by returning "false" where and
// when early exit from the overall pipeline is warranted and wanted.
type Processor[T any] func(item T) bool

// BatchProcessor is a function that processes a slice of [T] as part of a
// [Pipeline]. During execution, only copies of the original slice are given to
// the BatchProcessor, so it need not take care of copying the slice itself.
//
// It should return a success boolean and a slice (again, it need not be a copy)
// that is now holding the manipulations after processing. Any output the user
// needs to be aware of should be made with [Slog] calls, as no error type can
// be returned through the processor itself.
//
// During execution the pipeline should exit on the first failed processor, so
// "false" should be used only where pipeline failure is intended. The pipeline
// itself should not be context-aware, rather the BatchProcessor can capture a
// context and handle itself its cancellation by returning "false" where and
// when early exit from the overall pipeline is warranted and wanted.
type BatchProcessor[T any] func(items []T) ([]T, bool)

// Pipeline describes a structure that holds and executes [Processor] processor
// and [BatchProcessor] pre-/post-processor functions.
//
// During execution the pipeline processing functions should return "false" on
// the first failed processor. The pipeline itself should not be context-aware,
// rather the passed in [Processor] and [BatchProcessor] can capture a context
// and handle themselves its cancellation by returning "false" where and when
// early exit from the overall pipeline is warranted and wanted.
type Pipeline[T any] interface {
	// AddPreProcess adds a [BatchProcessor] pre-processor to the pipeline.
	AddPreProcess(processor BatchProcessor[T]) Pipeline[T]

	// Add adds a [Processor] to the pipeline.
	Add(processor Processor[T]) Pipeline[T]

	// AddPostProcess adds a [BatchProcessor] post-processor to the pipeline.
	AddPostProcess(processor BatchProcessor[T]) Pipeline[T]

	// PreProcess runs all [BatchProcessor] pre-processors on the given slice of
	// [T]. The function must be designed to operate on copies of the given
	// slice and also itself return a copy of the given slice (manipulated).
	PreProcess(items []T) ([]T, bool)

	// Process runs all [Processor] processors on the given [T].
	Process(item T) bool

	// PostProcess runs all [BatchProcessor] post-processors on the given slice
	// of [T]. The function must be designed to operate on copies of the given
	// slice and also itself return a copy of the given slice (manipulated).
	PostProcess(items []T) ([]T, bool)
}
