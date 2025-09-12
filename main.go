package main

import (
	"context"
	"database/sql"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var dsn = "root:rootpass@tcp(localhost:3306)/tests?parseTime=true"

func openConnect(driverName, dsn string) (*sql.DB, error) {
	return sql.Open(driverName, dsn)
}

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	logger := zap.NewNop()

	db, err := openConnect("mysql", dsn)
	if err != nil {
		panic(err)
	}
	repo := NewRepository(logger, db)

	err = repo.HAHAHA(ctx)
	if err != nil {
		panic(err)
	}
	/*
		migration := "db/migrations"

		parser := NewParser(migration)
		databases, err := parser.GetDatabasesFromMigrations(migration)
		if err != nil {
			panic(err)
		}

		for _, v := range databases {
			fmt.Println(v)
		}
	*/
}

// FileParser Интерфейс для возможного кастомного парсера.
type FileParser interface {
	GetDatabasesFromMigrations(migrationPath string) ([]Database, error)
}
