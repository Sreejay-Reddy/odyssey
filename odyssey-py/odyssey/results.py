class BuildLedgerResult:
    def __init__(
        self,
        key,
        targets,
        delegated,
        local
    ):
        self.key = key
        self.targets = targets
        self.delegated = delegated
        self.local = local

    @property
    def step_count(self):
        return len(self.targets)

class ExecuteResult:
    def __init__(
        self,
        key,
        target,
        success,
        status,
        response=None
    ):
        self.key = key
        self.target = target
        self.success = success
        self.status = status
        self.response = response
