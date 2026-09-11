package socket

import (
	"io"
	"fmt"
	"net"
	"errors"
	"encoding/binary"
	"encoding/json"

	"github.com/sreejay-reddy/odyssey/odyssey-agent/internal/registry"
)

func ReadHeader(conn net.Conn) ([]byte, error) {
    header := make([]byte, 4)

    _, err := io.ReadFull(conn, header)
	if err != nil {
        return nil, err
    }

    frameLength := binary.BigEndian.Uint32(header)
    buf := make([]byte, frameLength)

    return buf, nil
}

func DecodeRegistry(conn net.Conn, r *registry.Registry, buf []byte) error {
    if _, err := io.ReadFull(conn, buf); err != nil {
        return err
    }

    offset := 0

    if len(buf) < 32 {
        return errors.New("registry message too short")
    }

    var sdkID [16]byte
    var sessionID [16]byte

    copy(sdkID[:], buf[offset:offset+16])
    offset += 16

    copy(sessionID[:], buf[offset:offset+16])
    offset += 16

    if offset+4 > len(buf) {
        return errors.New("missing registry count")
    }

    count := binary.BigEndian.Uint32(buf[offset:])
    offset += 4

    for i := uint32(0); i < count; i++ {
        // TargetID
        if offset+4 > len(buf) {
            return errors.New("truncated target ID")
        }

        targetID := binary.BigEndian.Uint32(buf[offset:])
        offset += 4

        // Target name length
        if offset+2 > len(buf) {
            return errors.New("truncated target name length")
        }

        targetLength := int(binary.BigEndian.Uint16(buf[offset:]))
        offset += 2

        if offset+targetLength > len(buf) {
            return errors.New("truncated target name")
        }

        target := string(buf[offset : offset+targetLength])
        offset += targetLength

		ttlms := binary.BigEndian.Uint32(buf[offset:])
        offset += 4

        err := r.Add(registry.Registered{
            Target:       target,
            FunctionName: target,
            TargetID:     targetID,
			TTLMS: 		  ttlms,
        })
        if err != nil {
            return err
        }
    }

    return nil
}

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

		if offset+4 > len(buf) {
			return Result{}, fmt.Errorf("truncated result payload length")
		}

		resultLength := int(binary.BigEndian.Uint32(buf[offset:]))
		offset += 4

		if offset+resultLength > len(buf) {
			return Result{}, fmt.Errorf("truncated result payload")
		}

		executionResult := json.RawMessage(buf[offset : offset+resultLength])
		offset += resultLength

		result.Executions = append(result.Executions, ResultExecution{
			Key:      key,
			TargetID: targetID,
			ExecutionResult: executionResult,
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