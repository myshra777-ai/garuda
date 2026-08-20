// Copyright 2026 Rohit Mishra
// SPDX-License-Identifier: Apache-2.0

package noise

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// ComplexPipeline models a legacy async processing engine with anonymous types.
type ComplexPipeline struct {
	mu     sync.RWMutex
	Queue  chan func() error
	Config struct {
		WorkerCount int               `json:"worker_count"`
		Timeout     time.Duration     `json:"timeout"`
		Metadata    map[string]string `json:"metadata"`
	}
}

// HandlerFunc is an alias for nested functional transforms.
type HandlerFunc func(ctx context.Context, payload interface{}) (interface{}, error)

// NewComplexPipeline initializes the processing pipeline.
func NewComplexPipeline(workers int) *ComplexPipeline {
	p := &ComplexPipeline{
		Queue: make(chan func() error, workers*2),
	}
	p.Config.WorkerCount = workers
	p.Config.Timeout = 5 * time.Second
	return p
}

// Execute processes payloads using nested closures, defer-recovery, and type switching.
func (p *ComplexPipeline) Execute(ctx context.Context, input interface{}, transform HandlerFunc) (res interface{}, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in legacy pipeline: %v", r)
		}
	}()

	// Nested anonymous closure
	worker := func(innerCtx context.Context) (interface{}, error) {
		switch v := input.(type) {
		case string:
			return transform(innerCtx, v+"_processed")
		case []byte:
			return transform(innerCtx, string(v))
		default:
			return nil, fmt.Errorf("unsupported input type: %T", input)
		}
	}

	return worker(ctx)
}

// RawLegacyProcessor exposes an untyped interface for backward compatibility.
func RawLegacyProcessor(raw map[string]interface{}) (interface{}, error) {
	var result interface{} = raw["data"]
	return result, nil
}
