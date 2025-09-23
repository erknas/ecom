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
	migrationPath, envPath := fetchPaths()

	sourceURL := fmt.Sprintf("file://%s", migrationPath)
	databaseURL := fetchDatabaseURL(envPath, "POSTGRES_URL")

	m, err := migrate.New(sourceURL, databaseURL)
	if err != nil {
		panic(err)
	}

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Printf("no new migrations for %s", migrationPath)
			return
		}
		panic(err)
	}

	log.Printf("successful migration for %s", migrationPath)
}

func fetchPaths() (string, string) {
	var configPath, envPath string

	flag.StringVar(&configPath, "path", "", "path to migrations")
	flag.StringVar(&envPath, "env", "", "path to env file")
	flag.Parse()

	return configPath, envPath
}

func fetchDatabaseURL(envPath, env string) string {
	godotenv.Load(envPath)
	return os.Getenv(env)
}
