package batcher

import(
	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/transport/socket"
)

func BatchPostgres(msg socket.Message){
	for _, execution := range msg.Executions{
		print(execution)
	}
}