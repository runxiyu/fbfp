package main

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var db *sql.DB

func setup_database() error {
	var err error
	db, err = sql.Open(config.Db.Type, config.Db.Conn)
	if err != nil {
		return err
	} else {
		return nil
	}
}
