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

	return &Templater{logger: logger.Named("Templater")}
}

// Файл со сгенерированной моделью будет называться как оригинальное имя таблицы с суффиксом _model.go,
// В то время как имя структуры будет в CamelCase формате.
// А вот имя пакета нужно делать в соответствии с папкой, куда сохраняется файл.

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

func (t *Templater) ParseColumnsToFields(columns []model.Column) ([]Field, []CustomType) {
	fields := make([]Field, 0, len(columns))
	customTypes := make([]CustomType, 0, cap(columns))

	for _, column := range columns {
		if column.IsEnum() {
			t.logger.Debug("Column is enum type", zap.String("column", column.OriginalName))

			customTypes = append(customTypes, CustomType{
				Name:       column.CamelCaseName,
				ParentType: "string",
				Values:     column.EnumValues,
			})

			column.Type = column.CamelCaseName
		}

		field := Field{
			Name: column.CamelCaseName,
			Type: column.Type,
			Tags: t.GetTags(column),
		}
		fields = append(fields, field)
	}

	return fields, customTypes
}

func (t *Templater) GetTags(column model.Column) string {
	return `db:"` + column.OriginalName + `"`
}

func (t *Templater) CreateDBModel(database *model.Database, savePath string) error {
	t.logger.Info("Start creating model...", zap.String("database", database.TableNames.Original))
	fields, customTypes := t.ParseColumnsToFields(database.Columns)

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

const (
	templateText = `package {{.PackageName}} 

{{if .HasTimePackage}}import "time"
{{end}} 

type {{.ModelName}} struct {
{{- range .Fields}}
    {{.Name}} {{.Type}} ` + "`{{.Tags}}`" + `
{{- end}}
}

{{- range .CustomTypes}}
type {{.Name}} {{.ParentType}}

const (
{{- range .Values}}
    {{.Name}}{{.}} {{.Name}} = "{{.}}"
{{- end}}
)
{{end}}`
)
