import struct

from .protocol import Message, Execution

def decode_message(data):
    offset = 0

    version = data[offset]
    offset += 1

    message_type = data[offset]
    offset += 1

    flags = struct.unpack_from(">H", data, offset)[0]
    offset += 2

    batch_id = struct.unpack_from(">Q", data, offset)[0]
    offset += 8

    execution_count = struct.unpack_from(">I", data, offset)[0]
    offset += 4

    executions = []

    for _ in range(execution_count):
        key_length = struct.unpack_from(">H", data, offset)[0]
        offset += 2

        key = data[offset:offset + key_length].decode("utf-8")
        offset += key_length

        target_id = struct.unpack_from(">I", data, offset)[0]
        offset += 4

        input_length = struct.unpack_from(">I", data, offset)[0]
        offset += 4

        input_data = data[offset:offset + input_length]
        offset += input_length

        executions.append(
            Execution(
                key=key,
                target_id=target_id,
                input=input_data,
            )
        )

    return Message(
        version=version,
        type=message_type,
        flags=flags,
        batch_id=batch_id,
        executions=executions,
    )