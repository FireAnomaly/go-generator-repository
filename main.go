package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"

	"generatorFromMigrations/cli"
	"generatorFromMigrations/model"
	"generatorFromMigrations/parsers/mysql"
	"generatorFromMigrations/templater"
)

var (
	migrationPathInput = flag.String("in", "", "Path to the migration files")
	savePathInput      = flag.String("out", "", "Path to save generated models")
)

var (
	parser          FileParser
	tableManager    TableManager
	templateManager TemplaterManager
)

func main() {
	// logger, _ := zap.NewDevelopment()
	flag.Parse()
	if migrationPathInput == nil || *migrationPathInput == "" {
		log.Fatal("Migration path is required")
	}

	if savePathInput == nil || *savePathInput == "" {
		log.Fatal("Save path is required")
	}

	workDir, err := os.Getwd()
	if err != nil {
		log.Fatal("Failed to get working directory:", err)
	}

	savePath := workDir + "/" + *savePathInput
	migrationPath := workDir + "/" + *migrationPathInput

	fmt.Println("Migration Path:", migrationPath)
	fmt.Println("Save Path:", savePath)

	logger := zap.NewNop()

	parser = mysql.NewParser(migrationPath, logger)
	databases, err := parser.GetDatabasesFromMigrations(migrationPath)
	if err != nil {
		logger.Fatal("Failed to get migrations", zap.Error(err))
		panic(err)
	}

	// tableManager = cli.NewTableWriterOnCLI(logger)
	tableManager = cli.NewTview()
	err = tableManager.ManageTableByUser(databases)
	if err != nil {
		logger.Fatal("Failed to manage table by user", zap.Error(err))
		panic(err)
	}

	templateManager = templater.NewTemplater(logger)
	for _, db := range databases {
		err = templateManager.CreateDBModel(&db, savePath)
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
	CreateDBModel(database *model.Database, savePath string) error
}
