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

	var fields []Field
	customTypes := make(map[string]customType)
	hasCustomTypes := false
	isHaveTimePackage := false

	for _, column := range database.Columns {
		if column.IsEnum() {
			t.logger.Debug("Column is enum type", zap.String("column", column.OriginalName))
			hasCustomTypes = true
			customTypes[column.OriginalName] = customType{
				ParentType: "string",
				Values:     column.EnumValues,
			}

			column.Type = column.CamelCaseName
		}

		if column.IsTime() {
			isHaveTimePackage = true
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

	format := buildTemplate(&builderData{
		hasTimePackage: isHaveTimePackage,
		hasCustomTypes: hasCustomTypes,
		CustomTypes:    customTypes,
	})

	templ, err := template.New(database.TableNames.CamelCase).Parse(format)
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
	templateText = "package {{.PackageName}} \n \n"
)

type builderData struct {
	hasTimePackage bool
	hasCustomTypes bool
	CustomTypes    map[string]customType
}

type customType struct {
	ParentType string
	Values     []string
}

func buildTemplate(d *builderData) string {
	tmpl := templateText

	if d.hasTimePackage {
		tmpl += "import \"time\" \n \n"
	}

	tmpl += `type {{.ModelName}} struct {
{{- range .Fields}}
    {{.Name}} {{.Type}} ` + "`{{.Tags}}`" + `
{{- end}}
} ` + "\n"

	if d.hasCustomTypes {
		for typeName, typeInfo := range d.CustomTypes {
			tmpl += "type {{.typeName}} {{.ParentType}} \n\n" +
				"const (\n {{- range .Values}} \n {{}}" // fixme fuck this shit

			tmpl += "\ntype " + typeName + " " + typeInfo.ParentType + " \n\n"
			tmpl += "const (\n"
			for _, value := range typeInfo.Values {
				tmpl += "    " + typeName + value + " " + typeName + " = \"" + value + "\"\n"
			}
			tmpl += ")\n \n"

		}
	}

	return tmpl
}
