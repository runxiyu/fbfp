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

response_t: typing.TypeAlias = typing.Union[werkzeug.Response, flask.Response, str]
login_required_argument_t: typing.TypeAlias = typing.Callable[[], response_t]
login_required_t: typing.TypeAlias = typing.Callable[
    [login_required_argument_t], login_required_argument_t
]

VERSION = """fbfp v0.1

License: GNU Affero General Public License v3.0 or later
URL: https://sr.ht/~runxiyu/fbfp"""


def make_bp(login_required: login_required_t) -> flask.Blueprint:
    bp = flask.Blueprint("fbfp", __name__, url_prefix="/", template_folder="templates")

    @bp.route("/", methods=["GET"])
    @login_required
    def index() -> response_t:
        return flask.Response(flask.render_template("index.html"))

    return bp


def make_app(login_required: login_required_t) -> flask.App:
    app = flask.Flask(__name__)
    app.wsgi_app = werkzeug.middleware.proxy_fix.ProxyFix(  # type: ignore
        app.wsgi_app, x_for=1, x_proto=1, x_host=1, x_prefix=1
    )
    app.register_blueprint(make_bp(login_required), url_prefix="/")
    return app


if __name__ == "__main__":
    # This is obviously only for development.
    make_app(login_required=lambda i: i).run(port=8080, debug=True)
