package db

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

func Open() (*sql.DB, error) {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("error loading .env file")
	}

	var (
		pgHost     = os.Getenv("PGHOST")
		pgUser     = os.Getenv("PGUSER")
		pgPassword = os.Getenv("PGPASSWORD")
		pgDB       = os.Getenv("PGDATABASE")
		pgPort     = os.Getenv("PGPORT")
	)

	connStr := fmt.Sprintf("user=%s host=%s database=%s password=%s port=%s sslmode=disable", pgUser, pgHost, pgDB, pgPassword, pgPort)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	err = db.Ping()
	if err != nil {
		return nil, err
	}

	fmt.Println("database connected")

	return db, nil
}
