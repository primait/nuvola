import platform


def handler(event, context):
    return platform.node()
