package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"go.uber.org/zap"
)

func NewParser(migrationPath string, logger *zap.Logger) FileParser {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Parser{migrationPath: migrationPath, logger: logger}
}

type Parser struct {
	migrationPath string
	logger        *zap.Logger
}

func (p *Parser) GetDatabasesFromMigrations(migrationPath string) ([]Database, error) {
	p.logger.Debug("GetDatabasesFromMigrations called", zap.String("migrationPath", migrationPath))
	paths, err := p.GetPaths(migrationPath)
	if err != nil {
		p.logger.Debug("GetPaths error", zap.Error(err))
		return nil, err
	}

	var databases []Database
	for _, path := range paths {
		p.logger.Debug("Processing migration file", zap.String("path", path))
		var fileInfo []byte
		fileInfo, err = os.ReadFile(path)
		if err != nil {
			p.logger.Debug("os.ReadFile error", zap.Error(err), zap.String("path", path))
			return nil, err
		}

		TableName, err := p.GetTableName(fileInfo)
		if err != nil {
			p.logger.Debug("GetTableName error", zap.Error(err))
			return nil, err
		}
		p.logger.Debug("Parsed table name", zap.Any("TableName", TableName))

		columns, err := p.GetColumns(fileInfo)
		if err != nil {
			p.logger.Debug("GetColumns error", zap.Error(err))
			return nil, err
		}
		p.logger.Debug("Parsed columns", zap.Any("columns", columns))

		databases = append(databases, Database{
			TableNames: TableNames{
				CamelCase: TableName.CamelCase,
				Original:  TableName.Original,
			},
			Columns: columns,
		})
	}

	if len(databases) == 0 {
		p.logger.Debug("No databases found in migrations")
		return nil, ErrMigrationNotFound
	}

	p.logger.Debug("Successfully parsed databases", zap.Int("count", len(databases)))
	return databases, nil
}

func (p *Parser) GetColumns(fileInfo []byte) ([]Column, error) {
	p.logger.Debug("GetColumns called")

	lines := bytes.Lines((fileInfo)[1:])

	reColumn := regexp.MustCompile(`\b\w+\b`)
	reEnum := regexp.MustCompile(`'([^']*)'`)
	var columns []Column

	for line := range lines {
		line = bytes.Trim(bytes.TrimSpace(line), ",")

		var column Column
		matches := reColumn.FindAllSubmatch(line, -1)
		if len(matches) < 3 {
			continue
		}

		match := matches[0]

		column.Name = string(match[0]) // Преобразовывать в CamelCase для структуры, а для тега оставить как есть
		column.Type = string(match[1])

		isUint := false
		if bytes.Equal(bytes.ToLower(match[2]), []byte(`unsigned`)) {
			isUint = true
		}
		if isUint {
			column.Type = "uint"
		}

		if column.Type == "enum" {
			for _, v := range reEnum.FindAllSubmatch(line, -1) {
				column.EnumValues = append(column.EnumValues, string(v[1]))
			}
		}

		// todo добавить значения для default и isNull

		columns = append(columns, column)
	}

	p.logger.Debug("GetColumns finished", zap.Int("columnsCount", len(columns)))

	return columns, nil
}

func (p *Parser) GetTableName(file []byte) (TableNames, error) {
	p.logger.Debug("GetTableName called")
	patternTableName := `(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`
	reTableName, err := regexp.Compile(patternTableName)
	if err != nil {
		p.logger.Debug("regexp.Compile error", zap.Error(err))
		return TableNames{}, fmt.Errorf("failed get structure name: %w", err)
	}

	finded := reTableName.FindString(string(file))

	foundSlice := strings.Split(finded, " ")

	tableName := strings.Join(foundSlice[len(foundSlice)-1:], "")

	p.logger.Debug("Extracted table name", zap.String("tableName", tableName))

	return TableNames{
		CamelCase: p.toCamelCase(tableName),
		Original:  tableName,
	}, nil
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
	p.logger.Debug("GetPaths called", zap.String("migration", migration))
	pattern := fmt.Sprintf("%s/*.sql", migration)

	paths, err := filepath.Glob(pattern)
	if err != nil {
		p.logger.Debug("filepath.Glob error", zap.Error(err))
		return nil, fmt.Errorf("error finding migrations: %w", err)
	}

	if len(paths) == 0 {
		p.logger.Debug("No migration files found", zap.String("pattern", pattern))
		return nil, ErrMigrationNotFound
	}

	p.logger.Debug("Found migration files", zap.Int("count", len(paths)))
	return paths, nil
}
