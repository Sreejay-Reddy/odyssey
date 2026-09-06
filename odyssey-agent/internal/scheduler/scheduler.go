package scheduler

import (
	"fmt"
	"net"
)

type Scheduler struct {
	workers []Worker
	next    int
}

type Worker struct {
	ID   string
	Conn net.Conn
}

func NewScheduler(conns []net.Conn) *Scheduler {
	workers := make([]Worker, len(conns))

	for i, conn := range conns {
		workers[i] = Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Conn: conn,
		}
	}

	return &Scheduler{
		workers: workers,
	}
}

func (s *Scheduler) Next() Worker {
	worker := s.workers[s.next]
	s.next = (s.next + 1) % len(s.workers)
	return worker
}