package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	ErrMigrationNotFound = errors.New("migration not found")
	ErrInvalidMigration  = errors.New("migration is not valid")
)

func main() {
	migration := "db/migrations"
	paths, err := GetPaths(migration)
	if err != nil {
		panic(err)
	}

	patternTableName := `(?i)CREATE\s+TABLE\s+\w+`
	reTableName, err := regexp.Compile(patternTableName)
	if err != nil {
		panic(err)
	}

	for _, path := range paths {
		var fileInfo []byte
		fileInfo, err = os.ReadFile(path)
		if err != nil {
			panic(err)
		}

		finded := reTableName.FindString(string(fileInfo))

		tableName := strings.TrimLeft(finded, "CREATE TABLE ")
		unFormatedNames := strings.Split(tableName, "_")

		names := make([]string, 0, len(unFormatedNames))
		for _, v := range unFormatedNames {
			titleName := strings.ToTitle(v[:1])
			toCompileName := titleName + v[1:]
			names = append(names, toCompileName)
		}

		name := strings.Join(names, " ")
		fmt.Printf("struct name:%s \n", name)
	}

}

func GetPaths(migration string) ([]string, error) {
	pattern := fmt.Sprintf("%s/*.sql", migration)

	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding migrations: %w", err)
	}

	if len(paths) == 0 {
		return nil, ErrMigrationNotFound
	}

	fmt.Printf("Found %d migrations\n", len(paths))

	return paths, nil
}
