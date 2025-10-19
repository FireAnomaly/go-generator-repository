package main

import (
	"generatorFromMigrations/cli"
	"generatorFromMigrations/model"
	"generatorFromMigrations/parsers/mysql"
	"generatorFromMigrations/templater"
	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var (
	parser          FileParser
	tableManager    TableManager
	templateManager TemplaterManager
)

func main() {
	//logger, _ := zap.NewDevelopment()
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

	templateManager = templater.NewTemplater(logger)
	for _, db := range databases {
		err = templateManager.CreateDBModel(&db)
		if err != nil {
			logger.Fatal("Failed to create DB model", zap.Error(err))
			panic(err)
		}
	}
}

// FileParser Интерфейс для возможного кастомного парсера.
type FileParser interface {
	GetDatabasesFromMigrations(migrationPath string) ([]model.Database, error)
}

type TableManager interface {
	ManageTableByUser(dbs []model.Database) error
}

type TemplaterManager interface {
	CreateDBModel(database *model.Database) error
}
