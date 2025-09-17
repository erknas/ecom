package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/joho/godotenv"
)

func main() {
	path := fetchMigrationPath()

	sourceURL := fmt.Sprintf("file://%s", path)
	databaseURL := fetchDatabaseURL("POSTGRES_URL")

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("no new migrations for %s", path)
			return
		}
		panic(err)
	}

	log.Printf("successful migration for %s", path)
}

func fetchMigrationPath() string {
	var path string

	flag.StringVar(&path, "path", "", "path to the migrations")
	flag.Parse()

	return path
}

func fetchDatabaseURL(env string) string {
	if err := godotenv.Load(); err != nil {
		panic(err)
	}

	return os.Getenv(env)
}
