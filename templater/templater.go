package templater

import (
	"os"
	"strings"
	"text/template"

	"go.uber.org/zap"

	"github.com/FireAnomaly/go-generator-repository/model"
)

type Templater struct {
	logger *zap.Logger
}

func NewTemplater(logger *zap.Logger) *Templater {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &Templater{logger: logger.Named("Templater: ")}
}

type Field struct {
	Name string
	Type string
	Tags string
}

type CustomType struct {
	Name       string
	ParentType string
	Values     []string
}

func (t *Templater) parseColumnsToFields(camelCasedDBName string, columns []model.Column) ([]Field, []CustomType) {
	fields := make([]Field, 0, len(columns))
	customTypes := make([]CustomType, 0, cap(columns))

	for _, column := range columns {
		if column.IsEnum() {
			t.logger.Debug("Column is enum type", zap.String("column", column.OriginalName))

			customTypes = append(customTypes, CustomType{
				Name:       camelCasedDBName + column.CamelCaseName,
				ParentType: "string",
				Values:     column.EnumValues,
			})

			column.Type = camelCasedDBName + column.CamelCaseName
		}

		field := Field{
			Name: column.CamelCaseName,
			Type: column.Type,
			Tags: t.getTags(column),
		}
		fields = append(fields, field)
	}

	return fields, customTypes
}

func (t *Templater) getTags(column model.Column) string {
	return `db:"` + column.OriginalName + `"`
}

func (t *Templater) SaveModels(databases []*model.Database, savePath string) error {
	for _, db := range databases {
		if db == nil || db.Disabled {
			continue
		}

		if err := t.saveModel(db, savePath); err != nil {
			return err
		}
	}

	return nil
}

func (t *Templater) saveModel(database *model.Database, savePath string) error {
	t.logger.Info("Start creating model...", zap.String("database", database.TableNames.Original))
	fields, customTypes := t.parseColumnsToFields(database.TableNames.CamelCase, database.Columns)

	packageName := strings.Split(savePath, "/")[len(strings.Split(savePath, "/"))-1]

	data := struct {
		PackageName    string
		ModelName      string
		Fields         []Field
		HasTimePackage bool
		CustomTypes    []CustomType
	}{
		PackageName:    packageName,
		ModelName:      database.TableNames.CamelCase,
		Fields:         fields,
		HasTimePackage: database.IsHaveTime(),
		CustomTypes:    customTypes,
	}

	file, err := os.Create(savePath + "/" + database.TableNames.Original + "_model.go")
	if err != nil {
		t.logger.Error("Failed to create file", zap.Error(err))

		return err
	}

	templ, err := template.New(database.TableNames.CamelCase).Parse(templateText)
	if err != nil {
		t.logger.Error("Failed to parse template", zap.Error(err))

		return err
	}

	err = templ.Execute(file, data)
	if err != nil {
		t.logger.Error("Failed to execute template", zap.Error(err))

		return err
	}

	return nil
}

const templateText = `package {{.PackageName}} 
{{if .HasTimePackage}}
import "time"
{{end}} 
type {{.ModelName}} struct {
{{- range .Fields}}
    {{.Name}} {{.Type}} ` + "`{{.Tags}}`" + `
{{- end}}
}

{{- range .CustomTypes}}
type {{.Name}} {{.ParentType}}

const (
{{- $typeName := .Name}}
{{- range .Values}}
    {{$typeName}}{{.}} {{$typeName}} = "{{.}}"
{{- end}}
)
{{end}}
`
