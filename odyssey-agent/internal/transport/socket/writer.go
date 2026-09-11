package socket

import(
	"net"
	"context"
)

func RunWriter(ctx context.Context, conn net.Conn, send <-chan []byte) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case msg, ok := <-send:
			if !ok {
				return nil
			}

			if _, err := conn.Write(msg); err != nil {
				return err
			}
		}
	}
}