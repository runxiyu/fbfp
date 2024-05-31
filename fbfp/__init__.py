#!/usr/bin/env python3
# SPDX-License-Identifier: AGPL-3.0-or-later
# Copyright (c) 2024 Runxi Yu <https://runxiyu.org/>
# Upstream: https://git.sr.ht/~runxiyu/fbfp

from __future__ import annotations
from dataclasses import dataclass
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

import humanize
import jinja2
import flask
import werkzeug
import werkzeug.security
import werkzeug.middleware.proxy_fix
from sqlalchemy import select
from flask import url_for, Response, render_template, request

from .database import db
from . import models
from .types import *
from .exceptions import *


VERSION = """fbfp v0.1

License: GNU Affero General Public License v3.0 or later
URL: https://git.sr.ht/~runxiyu/fbfp"""


def no_login_required(
    f: typing.Callable[[context_t], response_t]
) -> typing.Callable[[], response_t]:
    @functools.wraps(f)
    def wrapper(*args: Any, **kwargs: Any) -> response_t:
        context = {
            "user": {
                "name": "Test User",
                "preferred_username": "test@example.org",
                "oid": "00000000-0000-0000-0000-000000000000",
            },
        }
        return f(context, *args, **kwargs)

    return wrapper


def ensure_user(context: context_t) -> models.User:
    if first_attempt := db.session.get(models.User, context["user"]["oid"]):
        # TODO: The following is done *just in case* the name or emails
        # change. This is obviously inefficient but shouldn't be a
        # bottleneck for now.
        first_attempt.name = context["user"]["name"]
        first_attempt.email = context["user"]["preferred_username"]
        return first_attempt
    user = models.User(
        oid=context["user"]["oid"],
        name=context["user"]["name"],
        email=context["user"]["preferred_username"],
    )
    db.session.add(user)
    db.session.commit()
    return user


def fbfpc_init(app: flask.Flask) -> None:
    max_request_size = app.config["FBFPC"]["max_request_size"]
    app.config["FBFPC"]["max_request_size_human"] = humanize.naturalsize(
        max_request_size, binary=True
    )
    max_file_size = app.config["FBFPC"]["max_file_size"]
    app.config["FBFPC"]["max_file_size_human"] = humanize.naturalsize(
        max_file_size, binary=True
    )
    app.config["FBFPC"]["upload_path"] = os.path.abspath(
        app.config["FBFPC"]["upload_path"]
    )


def fbfpc() -> typing.Any:
    return flask.current_app.config["FBFPC"]


def make_bp(login_required: login_required_t) -> flask.Blueprint:
    bp = flask.Blueprint("fbfp", __name__, url_prefix="/", template_folder="templates")

    @bp.route("/disclaimer", methods=["GET"])
    @login_required
    def disclaimer(context: context_t) -> response_t:
        user = ensure_user(context)
        return Response(render_template("disclaimer.html", user=user, fbfpc=fbfpc()))

    @bp.route("/", methods=["GET"])
    @login_required
    def index(context: context_t) -> response_t:
        user = ensure_user(context)
        wyours = user.works
        wothers = list(db.session.query(models.Work).filter(models.Work.user != user).filter(models.Work.public == True))  # type: ignore # FIXME
        return Response(
            render_template(
                "index.html", user=user, fbfpc=fbfpc(), wyours=wyours, wothers=wothers
            )
        )

    @bp.route("/nope", methods=["GET"])
    def test_nope() -> response_t:
        raise nope(418, "This endpoint tests the NOPE handler.")

    @bp.route("/static/<path:filename>", methods=["GET"])
    def static(filename: str) -> response_t:
        return flask.send_from_directory(fbfpc()["static_dir"], filename)

    # FIXME: Typing is broken because of how I typed login_required
    #        https://todo.sr.ht/~runxiyu/fbfp/6
    @bp.route("/work/<int:id>", methods=["GET"])  # type: ignore
    @login_required
    def work(context: context_t, id: int) -> response_t:
        user = ensure_user(context)
        if (not (work := db.session.get(models.Work, id))) or (
            (not work.public) and (work.user is not user)
        ):
            raise nope(404, "Work #%d does not exist or is private" % id)

        return Response(
            render_template("work.html", user=user, fbfpc=fbfpc(), work=work)
        )

    @bp.route("/work/<int:id>/update", methods=["POST"])  # type: ignore
    @login_required
    def work_update(context: context_t, id: int) -> response_t:
        user = ensure_user(context)

        if (not (work := db.session.get(models.Work, id))) or (work.user is not user):
            raise nope(403, "You cannot update nonexistant work or others' work.")

        work.anonymous = request.form.get("anonymous", None) != None
        work.public = request.form.get("public", None) != None
        work.active = request.form.get("active", None) != None

        db.session.flush()
        db.session.refresh(work)
        db.session.commit()

        flask.flash("Flags updated successfully.")

        return flask.redirect(url_for(".work", id=work.id))

    @bp.route("/work/<int:id>/edit", methods=["GET", "POST"])  # type: ignore # FIXME
    @login_required
    def work_edit(context: context_t, id: int) -> response_t:
        user = ensure_user(context)
        raise nope(501, "I haven't implemented this yet!")

    @bp.route("/work/<int:id>/delete", methods=["GET", "POST"])  # type: ignore
    @login_required
    def work_delete(context: context_t, id: int) -> response_t:
        user = ensure_user(context)
        if not ((work := db.session.get(models.Work, id)) and (work.user is user)):
            raise nope(
                403,
                "You cannot delete a nonexistent work or one owned by someone else.",
            )
        if request.method == "GET":
            return Response(
                render_template("work_delete.html", user=user, fbfpc=fbfpc(), work=work)
            )
        else:
            confirm = request.form.get("confirm", None) != None
            if not confirm:
                flask.flash(
                    "You must check the confirmation box for the deletion to commence."
                )
                return flask.redirect(url_for(".work_delete", id=work.id))
            if work.filename:
                try:
                    if not (
                        full_path := werkzeug.security.safe_join(
                            fbfpc()["upload_path"], work.filename
                        )
                    ):
                        raise FileNotFoundError
                    os.remove(full_path)
                except FileNotFoundError:
                    flask.flash("WARNING: file not found")
            db.session.delete(work)
            db.session.flush()
            db.session.commit()
            flask.flash("Deleted #%d" % id)
            return flask.redirect(url_for(".index"))

    @bp.route("/work/<int:wid>/comment/<int:cid>/delete", methods=["GET", "POST"])  # type: ignore # FIXME
    @login_required
    def work_comment_delete(context: context_t, wid: int, cid: int) -> response_t:
        user = ensure_user(context)
        raise nope(501, "I haven't implemented this yet!")
    @bp.route("/work/<int:wid>/comment/<int:cid>/edit", methods=["GET", "POST"])  # type: ignore # FIXME
    @login_required
    def work_comment_edit(context: context_t, wid: int, cid: int) -> response_t:
        user = ensure_user(context)
        raise nope(501, "I haven't implemented this yet!")

    @bp.route("/work/<int:wid>/comment/<int:cid>", methods=["GET", "POST"])  # type: ignore # FIXME
    @login_required
    def work_comment(context: context_t, wid: int, cid: int) -> response_t:
        user = ensure_user(context)
        if request.method == "POST":
            raise nope(501, "I haven't implemented this yet!")
        if (not (work := db.session.get(models.Work, wid))) or (
            (not work.public) and (work.user is not user)
        ):
            raise nope(404, "Work #%d does not exist or is private" % wid)
        if (
            (not (comment := db.session.get(models.WholeWorkComment, cid)))
            or (
                (not comment.public)
                and (comment.user is not user)
                and (work.user is not user)
            )
            or (comment.work is not work)
        ):
            raise nope(
                404,
                "Comment #%d does not exist, is private, or does not belong to this work"
                % wid,
            )
        return Response(
            render_template("work_comment.html", user=user, fbfpc=fbfpc(), work=work, comment=comment)
        )

    @bp.route("/work/<int:wid>/comment/new", methods=["GET", "POST"])  # type: ignore # FIXME
    @login_required
    def work_comment_new(context: context_t, wid: int) -> response_t:
        user = ensure_user(context)
        if (not (work := db.session.get(models.Work, wid))) or (
            (not work.public) and (work.user is not user)
        ):
            raise nope(404, "Work #%d does not exist or is private" % wid)
        if request.method == "GET":
            return Response(
                render_template(
                    "work_comment_new.html", user=user, fbfpc=fbfpc(), work=work
                )
            )

        form_file = request.files["file"]
        if filename := form_file.filename:
            if (
                shutil.disk_usage(fbfpc()["upload_path"]).free
                < fbfpc()["require_free_space"]
            ):
                raise nope(
                    500,
                    "The server does not have enough free space to safely store uploads.",
                )
            filename_base, filename_ext = os.path.splitext(os.path.basename(filename))
            with tempfile.NamedTemporaryFile(
                mode="w",
                suffix=filename_ext,
                prefix=filename_base + ".",
                dir=fbfpc()["upload_path"],
                delete=False,
            ) as fd:
                local_filename = fd.name
                form_file.save(local_filename)
        else:
            local_filename = None

        text: typing.Optional[str]

        try:
            title = request.form["title"]
            text = request.form["text"]
        except KeyError as e:
            raise nope(400, "Form does not include %s" % e.args[0])

        if not title.strip():
            raise nope(400, "You didn't include a title.")

        if not (text.strip()):
            if not local_filename:
                raise nope(
                    400,
                    "Your comment is basically empty. You need to upload a file or insert some text.",
                )
            text = None

        anonymous = request.form.get("anonymous", None) != None
        public = request.form.get("public", None) != None

        comment = models.WholeWorkComment(
            user=user,
            title=title,
            work=work,
            text=text,
            anonymous=anonymous,
            public=public,
            filename=os.path.basename(local_filename) if local_filename else None,
        )

        db.session.add(comment)
        db.session.flush()
        db.session.refresh(comment)
        db.session.commit()

        cid = comment.id
        assert type(cid) is int

        return flask.redirect(url_for(".work_comment", wid=work.id, cid=cid))

    @bp.route("/list", methods=["GET", "POST"])
    @login_required
    def list_(context: context_t) -> response_t:
        user = ensure_user(context)
        raise nope(501, "/list not implemented")

    @bp.route("/user/<oid>", methods=["GET"])  # type: ignore # FIXME
    @login_required
    def user(context: context_t, oid: str) -> response_t:
        user = ensure_user(context)
        if not (target := db.session.get(models.User, oid)):
            raise nope(404, "I don't know a user with the OID of %s" % oid)
        return {
            "oid": target.oid,
            "email": target.email,
            "name": target.name,
            "works": [
                {
                    "id": w.id,
                    "title": w.title,
                    "text": w.text,
                    "filename": w.filename,
                    "anonymous": w.anonymous,
                    "public": w.public,
                    "active": w.active,
                }
                for w in target.works
                if w.public or target is user  # TODO: Not sure if this is efficient
            ],
        }

    # Not authenticated because filename is partially random
    @bp.route("/file/<filename>", methods=["GET"])
    def file(filename: str) -> response_t:
        return flask.send_from_directory(
            fbfpc()["upload_path"], filename, as_attachment=True
        )

    @bp.route("/new", methods=["GET", "POST"])
    @login_required
    def new(context: context_t) -> response_t:
        user = ensure_user(context)
        if request.method == "GET":
            return Response(render_template("new.html", user=user, fbfpc=fbfpc()))
        form_file = request.files["file"]
        if filename := form_file.filename:
            if (
                shutil.disk_usage(fbfpc()["upload_path"]).free
                < fbfpc()["require_free_space"]
            ):
                raise nope(
                    500,
                    "The server does not have enough free space to safely store uploads.",
                )
            filename_base, filename_ext = os.path.splitext(os.path.basename(filename))
            with tempfile.NamedTemporaryFile(
                mode="w",
                suffix=filename_ext,
                prefix=filename_base + ".",
                dir=fbfpc()["upload_path"],
                delete=False,
            ) as fd:
                local_filename = fd.name
                form_file.save(local_filename)
        else:
            local_filename = None

        text: typing.Optional[str]

        try:
            title = request.form["title"]
            text = request.form["text"]
        except KeyError as e:
            raise nope(400, "Form does not include %s" % e.args[0])

        if not title.strip():
            raise nope(400, "You didn't include a title.")

        if not (text.strip()):
            if not local_filename:
                raise nope(
                    400,
                    "Your submission is basically empty. You need to upload a file or insert some text.",
                )
            text = None

        anonymous = request.form.get("anonymous", None) != None
        public = request.form.get("public", None) != None
        active = request.form.get("active", None) != None

        work = models.Work(
            user=user,
            title=title,
            text=text,
            anonymous=anonymous,
            active=active,
            public=public,
            filename=os.path.basename(local_filename) if local_filename else None,
        )

        db.session.add(work)
        db.session.flush()
        db.session.refresh(work)
        db.session.commit()

        id = work.id
        assert type(id) is int

        return flask.redirect(url_for(".work", id=id))

    return bp


def make_app(login_required: login_required_t, **config: typing.Any) -> flask.App:
    app = flask.Flask(__name__)
    app.wsgi_app = werkzeug.middleware.proxy_fix.ProxyFix(  # type: ignore
        app.wsgi_app, x_for=1, x_proto=1, x_host=1, x_prefix=1
    )
    app.register_blueprint(make_bp(login_required), url_prefix="/")
    app.config.update(**config)
    fbfpc_init(app)
    app.jinja_env.undefined = jinja2.StrictUndefined
    app.jinja_env.trim_blocks = True
    app.jinja_env.lstrip_blocks = True
    db.init_app(app)

    @app.errorhandler(nope)
    def handle_nope(
        exc: nope,
    ) -> response_t:
        tb = "".join(traceback.format_exception(exc, chain=True))
        return Response(
            render_template(
                "nope.html",
                msg=exc.args[1],
                error=tb,
                errver=VERSION,
                fbfpc=fbfpc(),
            ),
            status=exc.args[0],
        )

    with app.app_context():
        db.create_all()
    return app


def make_debug_app() -> flask.App:
    app = make_app(
        login_required=no_login_required,
        SQLALCHEMY_DATABASE_URI="sqlite:///test.db",
        FBFPC={
            "site_title": "FBFP Testing",
            "static_dir": "fbfp/static",
            "max_request_size": 3145728,  # not enforced here; should be enforced by nginx
            "max_file_size": 3000000,
            "upload_path": "uploads",
            "require_free_space": 3 * 1024 * 1024 * 1024,
            "version_info": VERSION,
        },
    )
    assert app.config["DEBUG"] == True
    app.config["SECRET_KEY"] = "DEBUG_ONLY"

    return app
