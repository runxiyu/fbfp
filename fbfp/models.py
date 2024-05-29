from sqlalchemy import Integer, String
from sqlalchemy.orm import mapped_column, relationship
from .database import Base
from typing import Optional


class User(Base):
    __tablename__ = "users"
    id = mapped_column(Integer, primary_key=True)
    name = mapped_column(String(50), unique=True)
    email = mapped_column(String(120), unique=True)

    def __init__(self, name: Optional[str] = None, email: Optional[str] = None) -> None:
        self.name = name
        self.email = email

    def __repr__(self) -> str:
        return f"<User {self.name!r}>"
