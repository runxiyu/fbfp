import flask
import functools

def login_required(f):
    @functools.wraps(f)
    def _(*args, **kwargs):
        return f(*args, **kwargs)
    return _
