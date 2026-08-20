package pipeline

import (
	"context"
	"fmt"
)

// TaskEngine processes pipeline jobs.
type TaskEngine struct {
	ID string
}

func (e *TaskEngine) RunTask(ctx context.Context, name string) error {
	fmt.Println("Running task:", name)
	return nil
}
