package main

import (
	"fmt"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type user_t struct {
	Subject  string `gorm:"primaryKey"`
	Name     string
	Email    string
	Sessions []session_t `gorm:"foreignKey:UserSubject"`
}

type session_t struct {
	Cookie      string `gorm:"primaryKey"`
	UserSubject string
	User        user_t `gorm:"foreignKey:UserSubject"`
}

func setup_database() error {
	var err error
	switch config.Db.Type {
	case "sqlite":
		db, err = gorm.Open(sqlite.Open(config.Db.Conn), &gorm.Config{})
		if err != nil {
			return err
		}
	default:
		return (fmt.Errorf("Database type \"%s\" unsupported", config.Db.Type))
	}
	return db.AutoMigrate(&user_t{}, &session_t{})
}
