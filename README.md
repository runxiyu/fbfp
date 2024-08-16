# Feedback {from,for} peers, a peer marking/commentary system

**This is completely unusable as of now. Check back later.**

At the end of the day it probably just targets the audience at YK Pao School
because I don't have time to maintain a project that could be used by others.

It is currently written in Python, but we're switching to Golang for
performance and dependency issues.

## Errors

* If you get `sqlalchemy.exc.ArgumentError: Class '<class 'fbfp.models.User'>' already has a primary mapper defined.`,
  then your Flask-SQLAlchemy is too old. Install according to `pyproject.toml`
  please.
