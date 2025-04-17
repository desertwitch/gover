package schema

// Processor is a function that processes [T] as part of a [Pipeline].
//
// It should return a success boolean, any output the user needs to be aware of
// should be made with [Slog] calls, as no error type can be returned through
// the processor itself.
type Processor[T any] func(item T) bool

// BatchProcessor is a function that processes a slice of [T] as part of a
// [Pipeline]. During execution, only copies of the given slice are operated on,
// so the processor function need not take care of copying the slice itself.
//
// It should return a success boolean and the copy of the original given slice
// that is now holding the manipulations after processing. Any output the user
// needs to be aware of should be made with [Slog] calls, as no error type can
// be returned through the processor itself.
type BatchProcessor[T any] func(items []T) ([]T, bool)

// Pipeline describes a structure that holds and executes [Processor] functions.
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
