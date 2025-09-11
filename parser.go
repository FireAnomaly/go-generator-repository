package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

func NewParser(migrationPath string) FileParser {
	return &Parser{migrationPath: migrationPath}
}

type Parser struct {
	migrationPath string
}

func (p *Parser) GetDatabasesFromMigrations(migrationPath string) ([]Database, error) {
	paths, err := p.GetPaths(migrationPath)
	if err != nil {
		return nil, err
	}

	var databases []Database
	for _, path := range paths {
		var fileInfo []byte
		fileInfo, err = os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		TableName, err := p.GetTableName(fileInfo)
		if err != nil {
			return nil, err
		}

		columns, err := p.GetColumns(fileInfo)
		if err != nil {
			return nil, err
		}

		databases = append(databases, Database{
			TableName: TableName,
			Columns:   columns,
		})
	}

	if len(databases) == 0 {
		return nil, ErrMigrationNotFound
	}

	return databases, nil
}

func (p *Parser) GetColumns(fileInfo []byte) ([]Column, error) {
	startColumns := strings.Index(string(fileInfo), "(")
	endColumns := strings.LastIndex(string(fileInfo), ")")

	rowsColumns := string(fileInfo)[startColumns+1 : endColumns]

	lines := strings.Split(rowsColumns, "\n")

	var columns []Column
	reColumn, err := regexp.Compile(`^\s*(\w+)\s+(\w+(?:\(\d+(?:,\s*\d+)?\))?)`)
	if err != nil {
		return nil, err
	}

	for _, line := range lines {
		line = strings.TrimSpace(line)
		line = strings.TrimSuffix(line, ",")

		matches := reColumn.FindStringSubmatch(line)
		if len(matches) < 3 {
			continue
		}

		column := Column{
			Name:   matches[1],
			Type:   matches[2],
			IsNull: !strings.Contains(line, "NOT NULL"),
		}

		columns = append(columns, column)
	}

	return columns, nil
}

func (p *Parser) GetTableName(file []byte) (string, error) {
	patternTableName := `(?i)CREATE\s+TABLE\s+\w+`
	reTableName, err := regexp.Compile(patternTableName)
	if err != nil {
		return "", fmt.Errorf("failed get structure name: %w", err)
	}

	finded := reTableName.FindString(string(file))

	tableName := strings.TrimLeft(finded, "CREATE TABLE ")

	return p.toCamelCase(tableName), nil
}

func (p *Parser) toCamelCase(snakeCase string) string {
	unFormatedNames := strings.Split(snakeCase, "_")

	names := make([]string, 0, len(unFormatedNames))
	for _, v := range unFormatedNames {
		titleName := strings.ToTitle(v[:1])
		toCompileName := titleName + v[1:]
		names = append(names, toCompileName)
	}

	return strings.Join(names, "")

}

func (p *Parser) GetPaths(migration string) ([]string, error) {
	pattern := fmt.Sprintf("%s/*.sql", migration)

	paths, err := filepath.Glob(pattern)
	if err != nil {
		return nil, fmt.Errorf("error finding migrations: %w", err)
	}

	if len(paths) == 0 {
		return nil, ErrMigrationNotFound
	}

	return paths, nil
}
