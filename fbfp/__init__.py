#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (c) 2024 Runxi Yu <https://runxiyu.org/>
# Upstream: https://git.sr.ht/~runxiyu/fbfp

from __future__ import annotations
import typing
import time
import os
import json
import sys
import traceback
import pathlib
import tempfile
import shutil
import datetime
import zoneinfo
import functools

import flask
import werkzeug
import werkzeug.security
import werkzeug.middleware.proxy_fix

context_t = typing.Optional[dict[str, typing.Any]]
response_t: typing.TypeAlias = typing.Union[werkzeug.Response, flask.Response, str]
login_required_t: typing.TypeAlias = typing.Callable[
    [typing.Callable[[context_t], response_t]], typing.Callable[[], response_t]
]

VERSION = """fbfp v0.1

License: GNU Affero General Public License v3.0 or later
URL: https://sr.ht/~runxiyu/fbfp"""


def no_login_required(
    f: typing.Callable[[context_t], response_t]
) -> typing.Callable[[], response_t]:
    @functools.wraps(f)
    def wrapper() -> response_t:
        context = {"user": {"name": "Test User 测试用户", "preferred_username": "test@example.org", "oid": "00000000-0000-0000-0000-000000000000"}}
        return f(context)

    return wrapper

def make_bp(login_required: login_required_t) -> flask.Blueprint:
    bp = flask.Blueprint("fbfp", __name__, url_prefix="/", template_folder="templates")

    @bp.route("/", methods=["GET"])
    @login_required
    def index(context: context_t) -> response_t:
        return flask.Response(flask.render_template("index.html"))

    return bp


def make_app(login_required: login_required_t) -> flask.App:
    app = flask.Flask(__name__)
    app.wsgi_app = werkzeug.middleware.proxy_fix.ProxyFix(  # type: ignore
        app.wsgi_app, x_for=1, x_proto=1, x_host=1, x_prefix=1
    )
    app.register_blueprint(make_bp(login_required), url_prefix="/")
    return app


def make_debug_app() -> flask.App:
    app = make_app(login_required=no_login_required)
    assert app.config["DEBUG"] == True
    return app


if __name__ == "__main__":
    make_debug_app().run(port=5000, debug=True)
