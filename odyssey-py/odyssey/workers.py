import asyncio
import socket
from multiprocessing import Process
from .utils import decode_message
from .execute import Execute
import time

class Worker:
    def __init__(self, worker_id, socket, registry):
        self.worker_id= worker_id
        self.socket = socket
        self.registry = registry

    async def recv_exactly(self, size):
        loop = asyncio.get_running_loop()
        data = bytearray()

        while len(data) < size:
            chunk = await loop.sock_recv(
                self.socket,
                size - len(data),
            )

            if not chunk:
                raise ConnectionError("agent disconnected")

            data.extend(chunk)

        return bytes(data)

    async def run(self):
        self.socket.setblocking(False)

        try:
            while True:
                header = await self.recv_exactly(4)
                size = int.from_bytes(header, "big")

                data = await self.recv_exactly(size)
                message = decode_message(data)

                for execution in message.executions:
                    asyncio.create_task(
                        Execute(
                            execution.key, 
                            execution.target_id,
                            execution.input,
                            self.registry
                        ).run()
                    )
        except asyncio.CancelledError:
            raise

        except ConnectionError:
            pass

        finally:
            self.socket.close()

    def worker_process(self):
        asyncio.run(self.run())

    def stop(self):
        self.socket.close()


class WorkerPool:
    def __init__(self, registry, workers):
        self.workers = workers
        self.registry = registry
        self.processes = []
        self.worker_instances = []

    def connect_worker(self, worker_id):
        path = f"/tmp/odyssey/{worker_id}.sock"

        while True:
            sock = socket.socket(socket.AF_UNIX, socket.SOCK_STREAM)

            try:
                sock.connect(path)
                return sock
            except ConnectionRefusedError:
                sock.close()
                time.sleep(1)


    def create_workers(self):
        for worker in range(self.workers):
            worker_id = f"worker-{worker}"

            sock = self.connect_worker(worker_id)

            worker_instance = Worker(
                worker_id=worker_id, 
                socket=sock,
                registry=self.registry
            )

            process = Process(
                target=worker_instance.worker_process,
            )
            process.start()

            self.worker_instances.append(worker_instance)
            self.processes.append(process)

            sock.close()

    def wait(self):
        for process in self.processes:
            process.join()

    def stop(self):
        for worker in self.worker_instances:
            worker.stop()

        for process in self.processes:
            if process.is_alive():
                process.terminate()

        for process in self.processes:
            process.join()

        self.worker_instances.clear()
        self.processes.clear()