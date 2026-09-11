package main

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/batcher"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/config"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/registry"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/scheduler"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/server"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/storage"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/storage/postgres"
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/transport/socket"
)

func runBatchLoop (ctx context.Context, 
	sch *scheduler.Scheduler, 
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

			registered, err := r.GetByName(claim.Target)
			if err != nil {
				return err
			}

			execute := socket.Execution{
				Key : claim.Key,
				TargetID : registered.TargetID,
				Input : claim.Input,
			}
			msg.Executions = append(msg.Executions, execute)
		}

		encodedMSG := socket.EncodeMessage(msg)

		worker.Send<- encodedMSG

	}
}

func runCompleteLoop (ctx context.Context, 
	resultconn net.Conn, 
	batchclient *batcher.Batcher,
	r *registry.Registry) (error) {
		for {
			header, err := socket.ReadHeader(resultconn)
			if err != nil {
				return err
			}

			result, err := socket.DecodeResult(header)
			if err != nil {
				return err
			}

			executions := make([]storage.Execution, 0, len(result.Executions))

			for _, execution := range result.Executions {

				registered, err := r.GetByID(execution.TargetID)
				if err != nil {
					return err
				}

				executed := storage.Execution {
					Key: execution.Key,
					Target: registered.Target,
					ExecutionResult: execution.ExecutionResult,
				}

				executions = append(executions, executed)
			}

			go batchclient.BatchComplete(ctx, executions)
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

	slog.Info("Database", "db_url", dbURL)

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

	slog.Info("connected")

	resultconn, err := socket.CreateResultSocket(ctx)
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

	sends := make([]chan<- []byte, len(workerconns))

	for i, conn := range workerconns {
		send := make(chan []byte, 64)

		go socket.RunWriter(ctx, conn, send)

		sends[i] = send
	}

	sch := scheduler.NewScheduler(workerconns, sends)

	pool, err := postgres.NewPool(ctx, cfg.Agent.Postgres, dbURL)
	if err != nil {
		return err
	}

	writer := postgres.New(pool, cfg)
	batchclient := batcher.New(writer, r, 64, time.Duration(1)*time.Second)

	go func(){
		err := runBatchLoop(ctx, sch, r, batchclient)
		if err != nil {
			stop()
		}
	}()

	go func(){
		err := runCompleteLoop(ctx, resultconn, batchclient, r)
		if err != nil {
			stop()
		}
	}()

	s := server.New(":8080")

	go func(){
		err := s.Start()
		if err != nil {
			s.Shutdown(ctx)
		}
	}()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(
		context.Background(),
		5*time.Second,
	)
	defer cancel()

	return s.Shutdown(shutdownCtx)
}

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}