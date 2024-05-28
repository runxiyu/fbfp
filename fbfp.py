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

import flask
import werkzeug
import werkzeug.security
import werkzeug.middleware.proxy_fix

# NOTE: The following import is *only* for development. When ready to deploy
# this to production, I should probably create a blueprint construction
# function as a closure (the constructor should accept a function that does
# the job of login_required). But that seems to require that I define all of the
# routes in the blueprint inside the closure, which is annoying for indentation.
# I guess I'll deal with that later.
from test_login import login_required

response_t: typing.TypeAlias = typing.Union[werkzeug.Response, flask.Response, str]

VERSION = """fbfp v0.1

License: GNU Affero General Public License v3.0 or later
URL: https://sr.ht/~runxiyu/fbfp"""

bp = flask.Blueprint("fbfp", __name__, url_prefix='/', template_folder="templates")

@bp.route("/", methods=["GET"])
@login_required
def index() -> response_t:
    return flask.Response(flask.render_template("index.html"))

def make_app() -> flask.App:
    app = flask.Flask(__name__)
    app.wsgi_app = werkzeug.middleware.proxy_fix.ProxyFix(  # type: ignore
        app.wsgi_app, x_for=1, x_proto=1, x_host=1, x_prefix=1
    )
    app.register_blueprint(bp, url_prefix='/')
    return app

if __name__ == "__main__":
    make_app().run(port=8080, debug=True)
