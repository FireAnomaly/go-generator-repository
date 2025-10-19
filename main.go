package main

import (
	"generatorFromMigrations/cli"
	"generatorFromMigrations/model"
	"generatorFromMigrations/parsers/mysql"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	parser       FileParser
	tableManager TableManager
)

func main() {
	// logger, _ := zap.NewDevelopment()
	logger := zap.NewNop()

	migration := "examples"

	parser = mysql.NewParser(migration, logger)
	databases, err := parser.GetDatabasesFromMigrations(migration)
	if err != nil {
		logger.Fatal("Failed to get migrations", zap.Error(err))
		panic(err)
	}

	tableManager = cli.NewTableWriterOnCLI(logger)
	err = tableManager.ManageTableByUser(databases)
	if err != nil {
		logger.Fatal("Failed to manage table by user", zap.Error(err))
		panic(err)
	}
}

// FileParser Интерфейс для возможного кастомного парсера.
type FileParser interface {
	GetDatabasesFromMigrations(migrationPath string) ([]model.Database, error)
}

type TableManager interface {
	ManageTableByUser(dbs []model.Database) error
}
