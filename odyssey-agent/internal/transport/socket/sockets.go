package socket

import(
	"net"
	"fmt"
	"time"
	"context"
	"path/filepath"
)

const(
	AckPath = "/tmp/odyssey-ack.sock"
	ResultPath = "/tmp/odyssey-result.sock"
)

const SocketDir = "/tmp/odyssey"


func createSocket(ctx context.Context, path string) (net.Conn, error) {
	var dialer net.Dialer

	for {
		conn, err := dialer.DialContext(ctx, "unix", path)
		if err == nil {
			return conn, nil
		}

		timer := time.NewTimer(time.Second)

		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()

		case <-timer.C:
		}
	}
}

func CreateWorkerSockets(ctx context.Context, workers int) ([]net.Conn, error) {
	sockets := make([]net.Conn, 0, workers)

	for worker := 0; worker < workers; worker++ {
		workerID := fmt.Sprintf("worker-%d", worker)
		path := filepath.Join(SocketDir, workerID+".sock")

		conn, err := createSocket(ctx, path)
		if err != nil {
			for _, socket := range sockets {
				socket.Close()
			}

			return nil, fmt.Errorf(
				"failed to connect worker socket %s: %w",
				workerID,
				err,
			)
		}

		sockets = append(sockets, conn)
	}

	return sockets, nil
}

func CreateAckSocket(ctx context.Context) (net.Conn, error) {
	return createSocket(ctx, AckPath)
}

func CreateResultSocket(ctx context.Context) (net.Conn, error) {
	return createSocket(ctx, ResultPath)
}