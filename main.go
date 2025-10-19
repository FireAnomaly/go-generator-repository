package main

import (
	"database/sql"

	_ "github.com/go-sql-driver/mysql"
	"go.uber.org/zap"
)

var dsn = "root:rootpass@tcp(localhost:3306)/tests?parseTime=true"

func openConnect(driverName, dsn string) (*sql.DB, error) {
	return sql.Open(driverName, dsn)
}

func main() {
	// ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	// defer cancel()

	logger, _ := zap.NewDevelopment()

	// db, err := openConnect("mysql", dsn)
	// if err != nil {
	// 	panic(err)
	// }
	// repo := NewRepository(logger, db)
	//
	// err = repo.HAHAHA(ctx)
	// if err != nil {
	// 	panic(err)
	// }

	migration := "examples"

	parser := NewParser(migration, logger)
	databases, err := parser.GetDatabasesFromMigrations(migration)
	if err != nil {
		logger.Fatal("Failed to get migrations", zap.Error(err))
		panic(err)
	}

	for _, v := range databases {
		logger.Info("TableNames:",
			zap.String("camel_cased", v.TableNames.CamelCase),
			zap.String("original", v.TableNames.Original))
		for _, v2 := range v.Columns {
			logger.Info("fields: ",
				zap.String("name", v2.OriginalName),
				zap.String("type", v2.Type),
				zap.Bool("is_null", v2.IsNull),
				zap.Any("default_value", v2.DefaultValue),
				zap.Strings("enum_values", v2.EnumValues),
			)
		}
	}
}

// FileParser Интерфейс для возможного кастомного парсера.
type FileParser interface {
	GetDatabasesFromMigrations(migrationPath string) ([]Database, error)
}
