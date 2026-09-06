class RegisteredTarget:
    def __init__(self, target, fn, target_id):
        self.target = target
        self.fn = fn
        self.function_name = fn.__name__
        self.target_id = target_id


class Register:
    def __init__(self, config):
        self.config = config
        self._registry = {}
        self._targets = {}
        self._next_id = 1

    def register(self, target, fn):
        registry = self.config.get("registry", {})
        default = registry.get("default")

        if not callable(fn):
            raise TypeError(
                "fn must be callable."
            )

        if target not in registry and default is None:
            raise ValueError(
                f"Target '{target}' is not defined in odyssey.yaml."
                "No default is defined in odyssey.yaml"
            )

        if target in self._registry:
            raise ValueError(
                f"Target '{target}' is already registered."
            )

        if not target.strip():
            raise ValueError(
                "target cannot be empty."
            )

        target_id = self._next_id
        self._next_id += 1

        registered = RegisteredTarget(target, fn, target_id)
        self._targets[target] = registered
        self._registry[target_id] = registered

        return registered

    def get(self, target_id):
        try:
            return self._registry[target_id]
        except KeyError:
            raise ValueError(
                f"Target '{target_id}' has not been registered."
            )

    def exists(self, target_id):
        return target_id in self._registry

    def targets(self):
        return list(self._registry.values())

    def registry(self):
        return dict(self._registry)

    def get_by_name(self, target):
        try:
            return self._targets[target]
        except KeyError:
            raise ValueError(
                f"Target '{target}' has not been registered."
            )

    def exists_by_name(self, target):
        return target in self._targets