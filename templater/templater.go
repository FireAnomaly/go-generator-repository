package templater

import (
	"os"
	"strings"
	"text/template"

	"go.uber.org/zap"

	"generatorFromMigrations/model"
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

var modelTemplate = `package {{.PackageName}}

type {{.ModelName}} struct {
{{- range .Fields}}
    {{.Name}} {{.Type}} ` + "`{{.Tags}}`" + `
{{- end}}
}
`

// Файл со сгенерированной моделью будет называться как оригинальное имя таблицы с суффиксом _model.go,
// В то время как имя структуры будет в CamelCase формате.
// А вот имя пакета нужно делать в соответствии с папкой, куда сохраняется файл.

type Field struct {
	Name string
	Type string
	Tags string
}

func (t *Templater) CreateDBModel(database *model.Database, savePath string) error {
	t.logger.Info("Start creating model...", zap.String("database", database.TableNames.Original))

	templ, err := template.New(database.TableNames.CamelCase).Parse(modelTemplate)
	if err != nil {
		t.logger.Error("Failed to parse template", zap.Error(err))

		return err
	}

	var fields []Field

	for _, column := range database.Columns {
		if column.IsEnum() {
			t.logger.Warn("Skipping enum column for model generation, so set him like string",
				zap.String("column", column.OriginalName))

			column.Type = "string"
		}

		field := Field{
			Name: column.CamelCaseName,
			Type: column.Type,
			Tags: `db:"` + column.OriginalName + `"`,
		}
		fields = append(fields, field)
	}

	packageName := strings.Split(savePath, "/")[len(strings.Split(savePath, "/"))-1]

	data := struct {
		PackageName string
		ModelName   string
		Fields      []Field
	}{
		PackageName: packageName,
		ModelName:   database.TableNames.CamelCase,
		Fields:      fields,
	}

	file, err := os.Create(savePath + "/" + database.TableNames.Original + "_model.go")
	if err != nil {
		t.logger.Error("Failed to create file", zap.Error(err))

		return err
	}

	err = templ.Execute(file, data)
	if err != nil {
		t.logger.Error("Failed to execute template", zap.Error(err))

		return err
	}

	return nil
}
