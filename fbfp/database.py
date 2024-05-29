from sqlalchemy import create_engine
from sqlalchemy.orm import scoped_session, sessionmaker, DeclarativeBase
import flask


class Base(DeclarativeBase):
    pass


def init_db() -> None:
    # import all modules here that might define models so that
    # they will be registered properly on the metadata.  Otherwise
    # you will have to import them first before calling init_db()
    engine = create_engine(flask.current_app.config["SQLALCHEMY_URL"])
    db_session = scoped_session(
        sessionmaker(autocommit=False, autoflush=False, bind=engine)
    )
    Base.query = db_session.query_property()
    from . import models

    Base.metadata.create_all(bind=engine)
