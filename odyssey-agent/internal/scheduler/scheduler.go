package scheduler

import (
	"fmt"
	"net"
	"sync"
)

type Scheduler struct {
	workers []Worker
	next    int
	mu 	 	sync.Mutex
}

type Worker struct {
	ID   string
	Conn net.Conn
	Send chan<- []byte
}

func NewScheduler(conns []net.Conn, sends []chan<- []byte) *Scheduler {
	workers := make([]Worker, len(conns))

	for i, conn := range conns {
		workers[i] = Worker{
			ID:   fmt.Sprintf("worker-%d", i),
			Conn: conn,
			Send: sends[i],
		}
	}

	return &Scheduler{
		workers: workers,
	}
}

func (s *Scheduler) Next() Worker {
	s.mu.Lock()
    defer s.mu.Unlock()

	worker := s.workers[s.next]
	s.next = (s.next + 1) % len(s.workers)
	
	return worker
}