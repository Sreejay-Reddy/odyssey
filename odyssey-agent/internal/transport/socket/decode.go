package socket

import (
	"encoding/binary"
	"fmt"
)

func DecodeResult(buf []byte) (Result, error) {
	const headerSize = 4 + 1 + 16 + 16 + 8 + 4

	if len(buf) < headerSize {
		return Result{}, fmt.Errorf("result too short")
	}

	frameLength := binary.BigEndian.Uint32(buf[:4])

	if int(frameLength) != len(buf)-4 {
		return Result{}, fmt.Errorf("invalid result frame length")
	}

	offset := 4

	result := Result{}

	result.Version = buf[offset]
	offset++

	copy(result.SDKID[:], buf[offset:offset+16])
	offset += 16

	copy(result.SessionID[:], buf[offset:offset+16])
	offset += 16

	result.BatchID = binary.BigEndian.Uint64(buf[offset:])
	offset += 8

	executionCount := binary.BigEndian.Uint32(buf[offset:])
	offset += 4

	result.Executions = make([]ResultExecution, 0, executionCount)

	for i := uint32(0); i < executionCount; i++ {
		if offset+2 > len(buf) {
			return Result{}, fmt.Errorf("truncated result key length")
		}

		keyLength := int(binary.BigEndian.Uint16(buf[offset:]))
		offset += 2

		if offset+keyLength > len(buf) {
			return Result{}, fmt.Errorf("truncated result key")
		}

		key := string(buf[offset : offset+keyLength])
		offset += keyLength

		if offset+4 > len(buf) {
			return Result{}, fmt.Errorf("truncated result target ID")
		}

		targetID := binary.BigEndian.Uint32(buf[offset:])
		offset += 4

		if offset+1 > len(buf) {
			return Result{}, fmt.Errorf("truncated result status")
		}

		status := ExecutionStatus(buf[offset])
		offset++

		result.Executions = append(result.Executions, ResultExecution{
			Key:      key,
			TargetID: targetID,
			Status:   status,
		})
	}

	if offset != len(buf) {
		return Result{}, fmt.Errorf(
			"unexpected trailing data: %d bytes",
			len(buf)-offset,
		)
	}

	return result, nil
}