from sqlalchemy import Integer, String, Boolean, ForeignKey
from sqlalchemy.orm import mapped_column, relationship, Mapped
from typing import Optional, List
from . import database


class User(database.db.Model):  # type: ignore
    __tablename__ = "fbfp_users"
    oid = mapped_column(String(36), primary_key=True)  # UUID
    name = mapped_column(String, unique=False)
    email = mapped_column(String, unique=True)
    can_submit = mapped_column(Boolean, unique=False)
    can_feedback = mapped_column(Boolean, unique=False)
    works: Mapped[List["Work"]] = relationship(back_populates="user")
    whole_work_comments: Mapped[List["WholeWorkComment"]] = relationship(
        back_populates="user"
    )

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


class Work(database.db.Model):  # type: ignore
    __tablename__ = "fbfp_works"
    __table_args__ = {"sqlite_autoincrement": True}
    id = mapped_column(Integer, primary_key=True)
    title = mapped_column(String, unique=False)
    text = mapped_column(String, unique=False)
    filename = mapped_column(String(255), unique=True)
    anonymous = mapped_column(Boolean, unique=False)
    public = mapped_column(Boolean, unique=False)
    active = mapped_column(Boolean, unique=False)
    oid = mapped_column(ForeignKey("fbfp_users.oid"))
    user: Mapped["User"] = relationship(back_populates="works")
    whole_work_comments: Mapped[List["WholeWorkComment"]] = relationship(
        back_populates="work"
    )

    def __init__(
        self,
        user: User,
        title: str,
        text: Optional[str],
        filename: Optional[str],
        anonymous: bool,
        public: bool,
        active: bool,
    ) -> None:
        self.user = user
        self.title = title
        if text:
            self.text = text
        else:
            self.text = None
        if filename:
            self.filename = filename
        else:
            self.filename = None
        self.anonymous, self.public, self.active = anonymous, public, active


class WholeWorkComment(database.db.Model):  # type: ignore
    __tablename__ = "fbfp_wwcomments"
    __table_args__ = {"sqlite_autoincrement": True}
    id = mapped_column(Integer, primary_key=True)
    wid = mapped_column(ForeignKey("fbfp_works.id"))
    work: Mapped["Work"] = relationship(back_populates="whole_work_comments")
    anonymous = mapped_column(Boolean, unique=False)
    public = mapped_column(Boolean, unique=False)
    filename = mapped_column(String(255), unique=True)
    text = mapped_column(String, unique=False)
    title = mapped_column(String, unique=False)
    oid = mapped_column(ForeignKey("fbfp_users.oid"))
    user: Mapped["User"] = relationship(back_populates="whole_work_comments")

    def __init__(
        self,
        user: User,
        title: str,
        work: Work,
        text: Optional[str],
        filename: Optional[str],
        anonymous: bool,
        public: bool,
    ) -> None:
        self.user = user
        self.title = title
        self.work = work
        if text:
            self.text = text
        else:
            self.text = None
        if filename:
            self.filename = filename
        else:
            self.filename = None
        self.anonymous, self.public = anonymous, public
