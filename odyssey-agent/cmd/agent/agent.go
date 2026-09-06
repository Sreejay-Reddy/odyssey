package main

import (
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/config"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/transport/socket"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/scheduler"
)

func run () (error) {
	path, err := config.FindYAML()
	if err != nil {
		return err
	}

	cfg, err := config.ReadYAML(path)
	if err != nil{
		return err
	}

	conns, err := socket.CreateWorkerSockets(cfg.Workers)
	if err != nil { 
		return err
	}

	sch := scheduler.NewScheduler(conns)
	return sch.Next().Conn.Close()
}