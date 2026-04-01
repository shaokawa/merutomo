package main

import (
	"database/sql"
	"errors"
	"log"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/pressly/goose/v3"
	"github.com/shaokawa/merutomo/backend/internal/config"
)

const migrationsDir = "migrations"

func main() {
	databaseURL, err := config.LoadDatabaseURL()
	if err != nil {
		log.Fatal(err)
	}

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := goose.SetDialect("postgres"); err != nil {
		log.Fatal(err)
	}

	command := "up"
	args := []string{}
	if len(os.Args) > 1 {
		command = os.Args[1]
		args = os.Args[2:]
	}

	if err := runCommand(db, command, args); err != nil {
		log.Fatal(err)
	}
}

func runCommand(db *sql.DB, command string, args []string) error {
	switch command {
	case "up":
		return goose.Up(db, migrationsDir)
	case "up-by-one":
		return goose.UpByOne(db, migrationsDir)
	case "down":
		return goose.Down(db, migrationsDir)
	case "down-to":
		if len(args) != 1 {
			return errors.New("down-to requires exactly one version argument")
		}
		return goose.DownTo(db, migrationsDir, mustParseVersion(args[0]))
	case "status":
		return goose.Status(db, migrationsDir)
	case "version":
		return goose.Version(db, migrationsDir)
	case "create":
		if len(args) != 2 {
			return errors.New("create requires name and type arguments, e.g. create add_posts sql")
		}
		return goose.Create(db, migrationsDir, args[0], args[1])
	default:
		return errors.New("unknown command: use up, up-by-one, down, down-to, status, version, or create")
	}
}

func mustParseVersion(raw string) int64 {
	version, err := goose.NumericComponent(raw)
	if err != nil {
		log.Fatal(err)
	}

	return version
}
