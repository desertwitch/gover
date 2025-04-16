package schema

// Processor is a function that processes [Moveable] as part of a [Pipeline].
//
// It should return a success boolean, any output the user needs to be aware of
// should be made with [slog] calls, as no error type can be returned through
// the processor itself.
type Processor func(*Moveable) bool

// BatchProcessor is a function that processes a slice of [Moveable] as part of
// a [Pipeline]. During execution, only copies of the given slice are operated
// on, so the processor function need not take care of copying the slice itself.
//
// It should return a success boolean and the copy of the original given slice
// that is now holding the manipulations after processing, any output the user
// needs to be aware of should be made with [slog] calls, as no error type can
// be returned through the processor itself.
type BatchProcessor func([]*Moveable) ([]*Moveable, bool)

// Pipeline describes a structure that holds and executes [Processor] functions.
type Pipeline interface {
	// AddPreProcess adds a [BatchProcessor] pre-processor to the pipeline.
	AddPreProcess(bp BatchProcessor) Pipeline

	// Add adds a [Processor] to the pipeline.
	Add(pr Processor) Pipeline

	// AddPostProcess adds a [BatchProcessor] post-processor to the pipeline.
	AddPostProcess(bp BatchProcessor) Pipeline

	// PreProcess runs all [BatchProcessor] pre-processors on the given slice of
	// [Moveable]. The function must be designed to operate on copies of the
	// given slice and also return a copy of the given slice (manipulated).
	PreProcess(moveables []*Moveable) ([]*Moveable, bool)

	// Process runs all [Processor] processors on the given [Moveable].
	Process(m *Moveable) bool

	// PostProcess runs all [BatchProcessor] post-processors on the given slice
	// of [Moveable]. The function must be designed to operate on copies of the
	// given slice and also return a copy of the given slice (manipulated).
	PostProcess(moveables []*Moveable) ([]*Moveable, bool)
}
