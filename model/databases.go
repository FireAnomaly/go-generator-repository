package model

import "errors"

var (
	ErrMigrationNotFound = errors.New("migration not found")
	ErrInvalidMigration  = errors.New("migration is not valid")
	ErrInvalidRegExp     = errors.New("invalid regexp")
)

type Database struct {
	TableNames         TableNames
	Columns            []Column
	FailedParseColumns []FailedParsedColumn
}

func (d *Database) IsHaveTime() bool {
	for _, col := range d.Columns {
		if col.IsTime() {
			return true
		}
	}

	return false
}

type TableNames struct {
	CamelCase string
	Original  string
}

type Column struct {
	OriginalName  string
	CamelCaseName string
	Type          string
	DefaultValue  any
	EnumValues    []string
	IsNull        bool
	// IsDisable     bool fixme // Нужно ли это?
}

func (c *Column) IsTime() bool {
	return c.Type == "time.Time"
}

func (c *Column) IsEnum() bool {
	return c.Type == "enum"
}

type FailedParsedColumn struct {
	OriginalName  string
	CamelCaseName string
	LineNumber    int
	Reason        error
}

// SupportedTypes содержит поддерживаемые типы данных и их синонимы - При парсинге приводить к нижнему регистру.
var SupportedTypes = map[string][]string{
	"int": {
		"int", "integer", "tinyint", "smallint", "mediumint", "bigint",
	},
	"uint": {
		"uint", "uint tinyint", "uint smallint", "uint mediumint", "uint int", "uint bigint",
		"int unsigned", "integer unsigned", "tinyint unsigned", "smallint unsigned",
		"mediumint unsigned", "bigint unsigned",
	},
	"float": {
		"float", "double", "decimal", "dec", "numeric",
		"float unsigned", "double unsigned", "decimal unsigned", "dec unsigned", "numeric unsigned",
	},
	"string": {
		"char", "varchar", "text", "tinytext", "mediumtext", "longtext",
		"set",
	},
	"enum": {
		"enum",
	},
	"bool": {
		"bool", "boolean",
	},
	"time.Time": {
		"date", "datetime", "timestamp", "time", "year",
	},
	"[]byte": {
		"blob", "tinyblob", "mediumblob", "longblob", "binary", "varbinary", "json",
	},
}

// ReverseSupportedTypes maps each synonym to its canonical type.
var ReverseSupportedTypes = map[string]string{
	"int":                "int",
	"integer":            "int",
	"tinyint":            "int",
	"smallint":           "int",
	"mediumint":          "int",
	"bigint":             "int",
	"uint":               "uint",
	"uint tinyint":       "uint",
	"uint smallint":      "uint",
	"uint mediumint":     "uint",
	"uint int":           "uint",
	"uint bigint":        "uint",
	"int unsigned":       "uint",
	"integer unsigned":   "uint",
	"tinyint unsigned":   "uint",
	"smallint unsigned":  "uint",
	"mediumint unsigned": "uint",
	"bigint unsigned":    "uint",
	"float":              "float",
	"double":             "float",
	"decimal":            "float",
	"dec":                "float",
	"numeric":            "float",
	"float unsigned":     "float",
	"double unsigned":    "float",
	"decimal unsigned":   "float",
	"dec unsigned":       "float",
	"numeric unsigned":   "float",
	"char":               "string",
	"varchar":            "string",
	"text":               "string",
	"tinytext":           "string",
	"mediumtext":         "string",
	"longtext":           "string",
	"enum":               "enum",
	"set":                "string",
	"bool":               "bool",
	"boolean":            "bool",
	"date":               "time.Time",
	"datetime":           "time.Time",
	"timestamp":          "time.Time",
	"time":               "time.Time",
	"year":               "time.Time",
	"blob":               "[]byte",
	"tinyblob":           "[]byte",
	"mediumblob":         "[]byte",
	"longblob":           "[]byte",
	"binary":             "[]byte",
	"varbinary":          "[]byte",
	"json":               "[]byte",
}

// Поведение при Enum
/* Нужно создавать 2 мапы:
Сервисная модель -> Модель репозитория
Сервисная модель <- Модель репозитория
Пока просто принимать создавать с типом string и
докинуть метод для валидации

*/
