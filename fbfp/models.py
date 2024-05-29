from sqlalchemy import Integer, String, Boolean
from sqlalchemy.orm import mapped_column, relationship
from typing import Optional
from . import database


class User(database.db.Model):  # type: ignore
    __tablename__ = "fbfp_users"
    oid = mapped_column(String(36), primary_key=True)  # UUID
    name = mapped_column(String(50), unique=False)
    email = mapped_column(String(120), unique=True)
    can_submit = mapped_column(Boolean, unique=False)
    can_feedback = mapped_column(Boolean, unique=False)

    def __init__(
        self,
        oid: str,
        email: str,
        name: str,
        can_submit: bool = True,
        can_feedback: bool = True,
    ) -> None:
        self.oid = oid
        self.name = name
        self.email = email
        self.can_submit = can_submit
        self.can_feedback = can_feedback

    def __repr__(self) -> str:
        return f"<User oid={self.oid!r} email={self.email!r}>"
