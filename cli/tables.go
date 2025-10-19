package cli

import (
	"os"

	"generatorFromMigrations/model"
	"github.com/fatih/color"
	"github.com/olekukonko/tablewriter"
	"github.com/olekukonko/tablewriter/renderer"
	"github.com/olekukonko/tablewriter/tw"
	"go.uber.org/zap"
)

type TableWriterOnCLI struct {
	logger *zap.Logger

	colorCfg renderer.ColorizedConfig
}

func NewTableWriterOnCLI(logger *zap.Logger) *TableWriterOnCLI {
	if logger == nil {
		logger = zap.NewNop()
	}

	// Configure colors: green headers, cyan/magenta rows, yellow footer
	colorCfg := renderer.ColorizedConfig{
		Header: renderer.Tint{
			FG: renderer.Colors{color.FgGreen, color.Bold}, // Green bold headers
			BG: renderer.Colors{color.BgHiWhite},
		},
		Column: renderer.Tint{
			FG: renderer.Colors{color.FgCyan}, // Default cyan for rows
			Columns: []renderer.Tint{
				{FG: renderer.Colors{color.FgMagenta}}, // Magenta for column 0
				{},                                     // Inherit default (cyan)
				{FG: renderer.Colors{color.FgHiRed}},   // High-intensity red for column 2
			},
		},
		Footer: renderer.Tint{
			FG: renderer.Colors{color.FgYellow, color.Bold}, // Yellow bold footer
			Columns: []renderer.Tint{
				{},                                      // Inherit default
				{FG: renderer.Colors{color.FgHiYellow}}, // High-intensity yellow for column 1
				{},                                      // Inherit default
			},
		},
		Border:    renderer.Tint{FG: renderer.Colors{color.FgWhite}}, // White borders
		Separator: renderer.Tint{FG: renderer.Colors{color.FgWhite}}, // White separators
	}

	return &TableWriterOnCLI{logger: logger.Named("CLI Table Writer: "), colorCfg: colorCfg}
}

var (
	tableHeadersForDatabases = []any{"Original Name", "CamelCased Name", "Number of Columns"}
	tableHeadersForColumns   = []any{"Column Name", "Data Type", "Is Null", "Default Value", "Enum Values"}
)

func (cli *TableWriterOnCLI) ManageTableByUser(dbs []model.Database) error {
	cli.logger.Debug("Start ManageTableByUser on CLI")
	for _, db := range dbs {
		if err := cli.WriteTableDatabase(db); err != nil {
			cli.logger.Error("Failed to write database table", zap.Error(err))

			return err
		}

		if err := cli.WriteTableColumns(db); err != nil {
			cli.logger.Error("Failed to write columns table", zap.Error(err))

			return err
		}
	}

	return nil
}

func (cli *TableWriterOnCLI) WriteTableDatabase(db model.Database) error {
	cli.logger.Debug("Start WriteTable on CLI for databases")

	table := cli.newTable()

	table.Header(tableHeadersForDatabases...)

	if err := table.Append([]any{
		db.TableNames.Original,
		db.TableNames.CamelCase,
		len(db.Columns),
	}); err != nil {
		cli.logger.Error("Failed to write table database", zap.Error(err))

		return err
	}

	if err := table.Render(); err != nil {
		cli.logger.Error("Failed to render table", zap.Error(err))

		return err
	}

	return nil
}

func (cli *TableWriterOnCLI) WriteTableColumns(db model.Database) error {
	cli.logger.Debug("Start WriteTable on CLI for columns", zap.String("table", db.TableNames.Original))

	table := cli.newTable()

	table.Header(tableHeadersForColumns...)

	for _, column := range db.Columns {
		if err := table.Append([]any{
			column.OriginalName,
			column.Type,
			column.IsNull,
			column.DefaultValue,
			column.EnumValues}); err != nil {
			cli.logger.Error("Failed to write table column",
				zap.String("table", db.TableNames.Original),
				zap.Error(err))

			return err
		}
	}

	if err := table.Render(); err != nil {
		cli.logger.Error("Failed to render table",
			zap.String("table", db.TableNames.Original),
			zap.Error(err))

		return err
	}

	return nil
}

func (cli *TableWriterOnCLI) newTable() *tablewriter.Table {
	return tablewriter.NewTable(os.Stdout,
		tablewriter.WithRenderer(renderer.NewColorized(cli.colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal}, // Wrap long content
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},     // Left-align rows
				ColMaxWidths: tw.CellWidth{Global: 25},
			},
			Footer: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignRight},
			},
		}),
	)
}
