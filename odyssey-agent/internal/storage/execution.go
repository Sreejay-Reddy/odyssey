package storage

import(
	"encoding/json"
)

type Execution struct {
	Key       string
	Target    string
	WorkerID string
	Status    string
	Attempts  int
	TTLMS 	  uint32
	Input     json.RawMessage
	ExecutionResult json.RawMessage
}