package socket

import(
	"net"
	"fmt"
	"path/filepath"
)

const(
	AckPath = "/tmp/odyssey-ack.sock"
	ResultPath = "/tmp/odyssey-result.sock"
)

const SocketDir = "/tmp/odyssey"


func createSocket(path string) (net.Conn, error){
	conn, err := net.Dial("unix", path)
	if err != nil{
		return nil, err
	}

	return conn, nil
}

func CreateWorkerSockets(workers int) ([]net.Conn, error) {
	sockets := make([]net.Conn, 0, workers)

	for worker := 0; worker < workers; worker++ {
		workerID := fmt.Sprintf("worker-%d", worker)
		path := filepath.Join(SocketDir, workerID+".sock")

		conn, err := createSocket(path)
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

func CreateAckSocket() (net.Conn, error) {
	return createSocket(AckPath)
}

func CreateResultSocket() (net.Conn, error) {
	return createSocket(ResultPath)
}