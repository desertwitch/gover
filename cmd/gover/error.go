package main

import "errors"

var (
	// ErrPipePreProcFailed occurs when a pre-processing pipeline has failed
	// during an operation.
	ErrPipePreProcFailed = errors.New("pre-processing pipeline has failed")

	// ErrPipePostProcFailed occurs when a post-processing pipeline has failed
	// during an operation.
	ErrPipePostProcFailed = errors.New("post-processing pipeline has failed")
)
