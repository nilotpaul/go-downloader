package config

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func initDB(DBURL string) (*sql.DB, error) {
	db, err := sql.Open("postgres", DBURL)
	if err != nil {
		return nil, err
	}

	if err = db.Ping(); err != nil {
		return nil, err
	}

	log.Println("DB Connected")

	return db, nil
}

func MustInitDB(DBURL string) *sql.DB {
	db, err := initDB(DBURL)
	if err != nil {
		panic(fmt.Sprintf("failed to connect to DB:  ", err))
	}

	return db
}
