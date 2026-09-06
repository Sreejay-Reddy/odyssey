from dataclasses import dataclass
from enum import IntEnum


PROTOCOL_VERSION = 1


class MessageType(IntEnum):
    SUBMIT = 1
    ACK = 2 
    RESULT = 3

class ExecutionStatus(IntEnum):
    SUCCESS = 1
    FAILED = 2


@dataclass(slots=True)
class Execution:
    key: str
    target_id: int
    input: bytes


@dataclass(slots=True)
class Message:
    version: int
    type: MessageType
    flags: int
    batch_id: int
    executions: list[Execution]