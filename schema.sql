CREATE TABLE users (
	id TEXT PRIMARY KEY NOT NULL,
	name TEXT,
	email TEXT
);
CREATE TABLE sessions (
	userid TEXT NOT NULL,
	cookie TEXT,
	expr INTEGER,
	FOREIGN KEY(userid) REFERENCES users(id)
);
