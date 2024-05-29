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

from . import make_debug_app

make_debug_app().run(port=5000, debug=True)
