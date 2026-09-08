package storage

import "context"

type Writer interface {
	Acquire(ctx context.Context, workerID string, limit int) ([]Execution, error)
	Complete(ctx context.Context, executions []Execution) error
}