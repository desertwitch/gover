package main

import "errors"

var (
	// ErrEvalPreProcFailed occurs when a pre-processing pipeline has failed
	// during evaluation.
	ErrEvalPreProcFailed = errors.New("pre-processing pipeline has failed")

	// ErrEvalPostProcFailed occurs when a post-processing pipeline has failed
	// during evaluation.
	ErrEvalPostProcFailed = errors.New("post-processing pipeline has failed")
)
