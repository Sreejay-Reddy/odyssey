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
	Command net.Conn
	Event net.Conn
	Send chan<- []byte
}

func NewScheduler(commands []net.Conn, events []net.Conn, sends []chan<- []byte) *Scheduler {
	workers := make([]Worker, len(commands))

    for i := range commands {
        workers[i] = Worker{
            ID:      fmt.Sprintf("worker-%d", i),
            Command: commands[i],
            Event:   events[i],
            Send:    sends[i],
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

func (s *Scheduler) Workers() []Worker {
	return s.workers
}