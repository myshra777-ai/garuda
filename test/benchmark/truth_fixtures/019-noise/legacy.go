package noise

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type ComplexPipeline struct {
	mu     sync.RWMutex
	Queue  chan func() error
	Config struct {
		WorkerCount int               `json:"worker_count"`
		Timeout     time.Duration     `json:"timeout"`
		Metadata    map[string]string `json:"metadata"`
	}
}

type HandlerFunc func(ctx context.Context, payload interface{}) (interface{}, error)

func NewComplexPipeline(workers int) *ComplexPipeline {
	p := &ComplexPipeline{
		Queue: make(chan func() error, workers*2),
	}
	p.Config.WorkerCount = workers
	p.Config.Timeout = 5 * time.Second
	return p
}

func (p *ComplexPipeline) Execute(ctx context.Context, input interface{}, transform HandlerFunc) (res interface{}, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("recovered from panic in legacy pipeline: %v", r)
		}
	}()

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

func RawLegacyProcessor(raw map[string]interface{}) (interface{}, error) {
	var result interface{} = raw["data"]
	return result, nil
}
