import kube.ctx
from util import ctx

def write_data(path: str, data: str):
    with open(path, "w") as file:
        file.write(data)


def get_landscape_config(cfg_name: str):
    factory = ctx().cfg_factory().cfg_set("hub")
    return factory.hub(cfg_name)


