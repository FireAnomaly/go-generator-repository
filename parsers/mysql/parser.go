package mysql

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/FireAnomaly/go-generator-repository/model"
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

	return &Parser{migrationPath: migrationPath, logger: logger.Named("MySQL Parser: ")}
}

func (p *Parser) GetDatabasesFromMigrations(migrationPath string) ([]model.Database, error) {
	p.logger.Info("Parse migrations", zap.String("migrationPath", migrationPath))
	paths, err := p.GetPaths(migrationPath)
	if err != nil {
		p.logger.Debug("GetPaths error", zap.Error(err))

		return nil, err
	}

	var databases []model.Database
	for _, path := range paths {
		p.logger.Info("Processing migration file", zap.String("path", path))
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

		columns, failedColumns, err := p.GetColumns(fileInfo)
		if err != nil {
			p.logger.Debug("GetColumns error", zap.Error(err))

			return nil, err
		}
		p.logger.Debug("Parsed columns")

		databases = append(databases, model.Database{
			TableNames: model.TableNames{
				CamelCase: TableName.CamelCase,
				Original:  TableName.Original,
			},
			Columns:            columns,
			FailedParseColumns: failedColumns,
		})
	}

	if len(databases) == 0 {
		p.logger.Debug("No databases found in migrations")

		return nil, model.ErrMigrationNotFound
	}

	p.logger.Info("Successfully parsed databases", zap.Int("count", len(databases)))

	return databases, nil
}

const (
	lenMatchesToParseNameAndType = 2
	minimumLineLengthToHaveUint  = 2
	minLenToEnums                = 2
)

var (
	reGetColumns = regexp.MustCompile(`\b\w+\b`)
	reGetEnums   = regexp.MustCompile(`\(([^)]*)\)`)
	reGetDefault = regexp.MustCompile(`(?i)DEFAULT\s+(['"]?[\w\s]*['"]?)`)
)

func (p *Parser) isLineContainsCreate(line []byte) bool {
	if bytes.Contains(bytes.ToUpper(line), []byte("CREATE TABLE")) {
		p.logger.Debug("Got CREATE TABLE line, skipping")

		return true
	}

	return false
}

func (p *Parser) clearLine(line []byte) []byte {
	return bytes.Trim(bytes.TrimSpace(line), ",")
}

func (p *Parser) GetColumns(fileInfo []byte) ([]model.Column, []model.FailedParsedColumn, error) {
	p.logger.Debug("GetColumns called")

	lines := bytes.Lines(fileInfo)

	var (
		columns       []model.Column
		currentLine   int
		failedColumns []model.FailedParsedColumn
	)

	for line := range lines {
		currentLine++
		p.logger.Debug("Processing line", zap.Int("lineNumber", currentLine))

		if p.isLineContainsCreate(line) {
			continue
		}

		line = p.clearLine(line)

		matches := reGetColumns.FindAllSubmatch(line, -1) // don't know how works this shit
		if len(matches) < lenMatchesToParseNameAndType {
			failedColumns = append(failedColumns, model.FailedParsedColumn{
				OriginalName:  "none",
				CamelCaseName: "none",
				LineNumber:    currentLine,
				Reason:        fmt.Errorf("line does not match expected column format"),
			})
			p.logger.Debug("Line does not match expected column format, skipping", zap.Int("lineNumber", currentLine))

			continue
		}

		originalName := string(matches[0][0])
		camelCaseName := p.toCamelCase(originalName)

		columnType, ok := model.ReverseSupportedTypes[string(bytes.ToLower(matches[1][0]))] // todo: rename
		if !ok {
			failedColumns = append(failedColumns, model.FailedParsedColumn{
				OriginalName:  originalName,
				CamelCaseName: camelCaseName,
				LineNumber:    currentLine,
				Reason:        fmt.Errorf("unsupported column type: %s", string(matches[1][0])),
			})
			p.logger.Debug("Unsupported column type found, skipping",
				zap.String("type", string(matches[1][0])),
				zap.Int("lineNumber", currentLine))

			continue
		}

		column := model.Column{
			OriginalName:  originalName,
			CamelCaseName: camelCaseName,
			Type:          columnType,
			IsNull:        !bytes.Contains(bytes.ToLower(line), []byte("not null")),
		}

		if strings.ToLower(column.Type) == "enum" {
			enums := reGetEnums.FindSubmatch(line)
			if len(enums) < minLenToEnums {
				failedColumns = append(failedColumns, model.FailedParsedColumn{
					OriginalName:  column.OriginalName,
					CamelCaseName: column.CamelCaseName,
					LineNumber:    currentLine,
					Reason:        fmt.Errorf("enum type found but no values present"),
				})
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
			if bytes.Equal(bytes.ToLower(matches[2][0]), []byte(`unsigned`)) {
				column.Type = "uint"
			}
		}

		if bytes.Contains(bytes.ToLower(line), []byte("default")) {
			matchesDefaults := reGetDefault.FindSubmatch(line)
			if len(matchesDefaults) > 1 {
				p.logger.Debug("Found default value", zap.String("value", string(matchesDefaults[1])))
				column.DefaultValue = strings.Trim(string(matchesDefaults[1]), `'`)
			}
		}

		columns = append(columns, column)
	}

	p.logger.Debug("GetColumns finished", zap.Int("columnsCount", len(columns)))

	return columns, failedColumns, nil
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
