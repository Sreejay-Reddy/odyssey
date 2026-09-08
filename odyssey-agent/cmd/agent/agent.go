package main

import (
	"os/signal"
	"syscall"
	"context"
	"errors"
	"time"
	"os"
	

	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/batcher"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/config"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/registry"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/scheduler"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/storage/postgres"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/transport/socket"
)

func runBatchLoop (ctx context.Context, 
	sch *scheduler.Scheduler, 
	postgres *postgres.Writer,
	r *registry.Registry, 
	batchclient *batcher.Batcher) (error) {
	for {
		worker := sch.Next()

		batch, err := batchclient.Next(ctx, worker.ID)
		if err != nil {
			return  err
		}

		msg := socket.Message{
			Type : socket.MessageSubmit,
			Version : socket.ProtocolVersion,
			Flags : 0,
			BatchID : batch.ID,
			Executions: make([]socket.Execution, 0, len(batch.Executions)),
		}

		for _, claim := range batch.Executions {

			registred, err := r.GetByName(claim.Target)
			if err != nil {
				return err
			}

			execute := socket.Execution{
				Key : claim.Key,
				TargetID : registred.TargetID,
				Input : claim.Input,
			}
			msg.Executions = append(msg.Executions, execute)
		}

		encodedMSG := socket.EncodeMessage(msg)

		_, err = worker.Conn.Write(encodedMSG)
		if err != nil {
			return  err
		}

	}
}

func run () (error) {
    ctx, stop := signal.NotifyContext(
        context.Background(),
        os.Interrupt,
        syscall.SIGTERM,
    )
    defer stop()

	dbURL, exists := os.LookupEnv("DATABASE_URL")
	if exists != true{
		return errors.New("DATABASE_URL not set")
	}

	path, err := config.FindConfig()
	if err != nil {
		return err
	}

	cfg, err := config.LoadConfigYAML(path)
	if err != nil{
		return err
	}

	ackconn, err := socket.CreateAckSocket(ctx)
	if err != nil {
		return err
	}

	msgSize, err := socket.ReadHeader(ackconn)
	if err != nil {
		return err
	}

	r := registry.New()

	err = socket.DecodeRegistry(ackconn, r, msgSize)
	if err != nil {
		return err
	}

	workerconns, err := socket.CreateWorkerSockets(ctx, cfg.Agent.SDK.Workers)
	if err != nil { 
		return err
	}

	sch := scheduler.NewScheduler(workerconns)

	pool, err := postgres.NewPool(ctx, cfg.Agent.Postgres, dbURL)
	if err != nil {
		return err
	}

	writer := postgres.New(pool, cfg)
	batchclient := batcher.New(writer,	64, time.Duration(1)*time.Second)

	go func(){
		err := runBatchLoop(ctx, sch, writer, r, batchclient)
		if err != nil {
			stop()
		}
	}()

	<-ctx.Done()

	return nil
}