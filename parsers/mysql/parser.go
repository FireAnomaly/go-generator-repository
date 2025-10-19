package mysql

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"generatorFromMigrations/model"
	"go.uber.org/zap"
)

type Parser struct {
	migrationPath string
	logger        *zap.Logger
}

func NewParser(migrationPath string, logger *zap.Logger) *Parser {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Parser{migrationPath: migrationPath, logger: logger}
}

func (p *Parser) GetDatabasesFromMigrations(migrationPath string) ([]model.Database, error) {
	p.logger.Debug("GetDatabasesFromMigrations called", zap.String("migrationPath", migrationPath))
	paths, err := p.GetPaths(migrationPath)
	if err != nil {
		p.logger.Debug("GetPaths error", zap.Error(err))

		return nil, err
	}

	var databases []model.Database
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

		databases = append(databases, model.Database{
			TableNames: model.TableNames{
				CamelCase: TableName.CamelCase,
				Original:  TableName.Original,
			},
			Columns: columns,
		})
	}

	if len(databases) == 0 {
		p.logger.Debug("No databases found in migrations")

		return nil, model.ErrMigrationNotFound
	}

	p.logger.Debug("Successfully parsed databases", zap.Int("count", len(databases)))

	return databases, nil
}

const (
	lenMatchesToParseNameAndType = 2
	minimumLineLengthToHaveUint  = 2
	minLenToEnums                = 2
)

func (p *Parser) GetColumns(fileInfo []byte) ([]model.Column, error) {
	p.logger.Debug("GetColumns called")

	lines := bytes.Lines(fileInfo)

	reColumn := regexp.MustCompile(`\b\w+\b`)
	reEnum := regexp.MustCompile(`\(([^)]*)\)`)
	reDefault := regexp.MustCompile(`(?i)DEFAULT\s+(['"]?[\w\s]*['"]?)`)

	var columns []model.Column
	var currentLine int

	for line := range lines {
		currentLine++
		if bytes.Contains(bytes.ToUpper(line), []byte("CREATE TABLE")) {
			p.logger.Debug("Got CREATE TABLE line, skipping", zap.Int("lineNumber", currentLine))

			continue
		}

		p.logger.Debug("Processing line", zap.Int("lineNumber", currentLine), zap.ByteString("lineContent", line))

		line = bytes.Trim(bytes.TrimSpace(line), ",")

		p.logger.Debug("line after trim", zap.ByteString("trimmedLine", line))

		var column model.Column
		matches := reColumn.FindAllSubmatch(line, -1)
		if len(matches) < lenMatchesToParseNameAndType {
			p.logger.Warn("Line does not match expected column format, skipping", zap.Int("lineNumber", currentLine))

			continue
		}

		columnType, ok := model.ReverseSupportedTypes[string(bytes.ToLower(matches[1][0]))]
		if !ok {
			p.logger.Error("Unsupported column type found, skipping",
				zap.String("type", string(matches[1][0])),
				zap.Int("lineNumber", currentLine))

			continue
		}

		column.OriginalName = string(matches[0][0])
		column.CamelCaseName = p.toCamelCase(column.OriginalName)
		column.Type = columnType
		column.IsNull = true

		if bytes.Contains(bytes.ToLower(line), []byte("not null")) {
			column.IsNull = false
		}

		if strings.ToLower(column.Type) == "enum" {
			enums := reEnum.FindSubmatch(line)
			if len(enums) < minLenToEnums {
				p.logger.Debug("Enum type found but no values present, skipping",
					zap.String("column", column.OriginalName),
					zap.Int("lineNumber", currentLine))

				continue
			}

			enumValues := strings.Split(string(enums[1]), ",")
			for _, enumValue := range enumValues {
				trimmedEnumValue := strings.TrimPrefix(strings.Trim(enumValue, `'`), ` '`)
				p.logger.Debug("Found enum", zap.String("enum", trimmedEnumValue))

				column.EnumValues = append(column.EnumValues, trimmedEnumValue)
			}
		}

		// Проверка на unsigned для целочисленных типов
		if len(matches) > minimumLineLengthToHaveUint {
			isUint := false
			if bytes.Equal(bytes.ToLower(matches[2][0]), []byte(`unsigned`)) {
				isUint = true
			}
			if isUint {
				column.Type = "uint"
			}
		}

		if bytes.Contains(bytes.ToLower(line), []byte("default")) {
			matchesDefaults := reDefault.FindSubmatch(line)
			if len(matchesDefaults) > 1 {
				p.logger.Debug("Found default value", zap.String("value", string(matchesDefaults[1])))
				column.DefaultValue = strings.Trim(string(matchesDefaults[1]), `'`)
			}
		}

		columns = append(columns, column)
	}

	p.logger.Debug("GetColumns finished", zap.Int("columnsCount", len(columns)))

	return columns, nil
}

func (p *Parser) GetTableName(file []byte) (model.TableNames, error) {
	p.logger.Debug("GetTableName called")
	patternTableName := `(?i)CREATE\s+TABLE\s+(?:IF\s+NOT\s+EXISTS\s+)?(\w+)`
	reTableName, err := regexp.Compile(patternTableName)
	if err != nil {
		p.logger.Debug("regexp.Compile error", zap.Error(err))

		return model.TableNames{}, fmt.Errorf("failed get structure name: %w", err)
	}

	finded := reTableName.FindString(string(file))

	foundSlice := strings.Split(finded, " ")

	tableName := strings.Join(foundSlice[len(foundSlice)-1:], "")

	p.logger.Debug("Extracted table name", zap.String("tableName", tableName))

	return model.TableNames{
		CamelCase: p.toCamelCase(tableName),
		Original:  tableName,
	}, nil
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

		return nil, model.ErrMigrationNotFound
	}

	p.logger.Debug("Found migration files", zap.Int("count", len(paths)))

	return paths, nil
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
