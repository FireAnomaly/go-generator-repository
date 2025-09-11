package main

import (
	"fmt"
)

func main() {
	migration := "db/migrations"

	parser := NewParser(migration)
	databases, err := parser.GetDatabasesFromMigrations(migration)
	if err != nil {
		panic(err)
	}

	for _, v := range databases {
		fmt.Println(v)
	}
}

// FileParser Интерфейс для возможного кастомного парсера.
type FileParser interface {
	GetDatabasesFromMigrations(migrationPath string) ([]Database, error)
}
