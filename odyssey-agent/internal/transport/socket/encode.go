package socket

import "encoding/binary"

var byteOrder = binary.BigEndian

func EncodeMessage(msg Message) []byte {
	size := messageSize(msg)
	buf := make([]byte, size)

	offset := 4 

	buf[offset] = msg.Version
	offset++

	buf[offset] = byte(msg.Type)
	offset++

	byteOrder.PutUint16(buf[offset:], msg.Flags)
	offset += 2

	byteOrder.PutUint64(buf[offset:], msg.BatchID)
	offset += 8

	byteOrder.PutUint32(buf[offset:], uint32(len(msg.Executions)))
	offset += 4

	for _, execution := range msg.Executions {
		key := []byte(execution.Key)

		byteOrder.PutUint16(buf[offset:], uint16(len(key)))
		offset += 2

		copy(buf[offset:], key)
		offset += len(key)

		byteOrder.PutUint32(buf[offset:], execution.TargetID)
		offset += 4

		byteOrder.PutUint32(buf[offset:], uint32(len(execution.Input)))
		offset += 4

		copy(buf[offset:], execution.Input)
		offset += len(execution.Input)
	}

	byteOrder.PutUint32(buf[:4], uint32(size-4))

	return buf
}

func EncodeAck(ack Ack) []byte {
	const size = 4 + 1 + 8 + 16 + 16

	buf := make([]byte, size)

	offset := 4 

	buf[offset] = ack.Version
	offset++

	byteOrder.PutUint64(buf[offset:], ack.BatchID)
	offset += 8

	copy(buf[offset:], ack.SessionID[:])
	offset += 16

	copy(buf[offset:], ack.SDKID[:])

	byteOrder.PutUint32(buf[:4], uint32(size-4))

	return buf
}

func messageSize(msg Message) int {
	size := 4 + 1 + 1 + 2 + 8 + 4

	for _, execution := range msg.Executions {
		size += 2                    // key length
		size += len(execution.Key)   // key
		size += 4                    // target ID
		size += 4                    // input length
		size += len(execution.Input) // input
	}

	return size
}