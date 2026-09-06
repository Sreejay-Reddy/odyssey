import inspect
import asyncio

class Execute:
    def __init__(
        self,
        key,
        target_id,
        input,
        registry
    ):

        self.key = key
        self.target_id = target_id
        self.input = input
        self.registry = registry

    async def run(self):

        registered = self.registry.get(self.target_id)
        if registered is None:
            raise RuntimeError(
                f"Unknown target: {self.target_id}"
            )
        
        fn = registered.fn
        kwargs = self.input or {}
                
        try:

            if inspect.iscoroutinefunction(fn):
                return await fn(**kwargs)
            else:
                loop = asyncio.get_running_loop()

                return await loop.run_in_executor(
                    None,
                    lambda: fn(**kwargs),
                )

        except Exception as e:
            raise RuntimeError("Could not execute function") from e

        
