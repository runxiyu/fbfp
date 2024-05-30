from typing import *
import flask
import werkzeug

context_t = dict[str, Any]
response_t: TypeAlias = Union[
    werkzeug.Response, flask.Response, str, dict[str, Any], list[Any]
]
login_required_t: TypeAlias = Callable[
    [Callable[[context_t], response_t]], Callable[[], response_t]
]
