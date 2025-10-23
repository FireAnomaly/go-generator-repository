package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strconv"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/FireAnomaly/go-generator-repository/model"
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
			BG: renderer.Colors{color.BgBlack},
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
	tableHeadersForFailedColumns = []any{"Original Name", "CamelCased Name", "Line Number", "Reason"}
	tableHeadersForColumns       = []any{"Column Name", "Data Type", "Is Null", "Default Value", "Enum Values"}
)

func (cli *TableWriterOnCLI) ManageTableByUser(dbs []model.Database) error {
	cli.logger.Debug("Start ManageTableByUser on CLI")

	databases := map[string]model.Database{}
	for _, db := range dbs {
		databases[db.TableNames.Original] = db
	}

	cli.UserGetDB(dbs)

	panic("stop")

}

var clearConsole map[string]func() // create a map for storing clearConsole funcs

func init() {
	clearConsole = make(map[string]func()) // Initialize it
	clearConsole["linux"] = func() {
		cmd := exec.Command("clear") // Linux example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
	clearConsole["windows"] = func() {
		cmd := exec.Command("cmd", "/c", "cls") // Windows example, its tested
		cmd.Stdout = os.Stdout
		cmd.Run()
	}
}

func CallClearConsole() {
	value, ok := clearConsole[runtime.GOOS] // runtime.GOOS -> linux, windows, darwin etc.
	if ok {                                 // if we defined a clearConsole func for that platyform:
		value() // we execute it
	} else { // unsupported platform
		panic("Your platform is unsupported! I can't clearConsole terminal screen :(")
	}
}

func (cli *TableWriterOnCLI) UserGetDB(dbs []model.Database) {
	CallClearConsole()
	cli.writeTable(dbs)
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		CallClearConsole()
		switch key.Code {
		case keys.Down:
			cli.writeTable(dbs)
			return false, nil
		case keys.Up:
			cli.writeTable(dbs)
			return false, nil
		case keys.CtrlC:
			CallClearConsole()
			return true, nil
		}

		return false, nil
	})
}

func (cli *TableWriterOnCLI) writeTable(dbs []model.Database) {
	cli.logger.Debug("Start WriteTable on CLI for databases")

	t := cli.setTable()
	t.Header([]string{"Num", "Original Name", "CamelCased Name", "Number of Columns"})

	num := 1
	for _, v := range dbs {
		if err := t.Append([]string{
			strconv.Itoa(num),
			v.TableNames.Original,
			v.TableNames.CamelCase,
			strconv.Itoa(len(v.Columns)),
		}); err != nil {
			cli.logger.Error("Failed to write t database", zap.Error(err))
		}

		num++
	}

	t.Render()

	return
}

func (cli *TableWriterOnCLI) WriteTableColumns(db model.Database) error {
	cli.logger.Debug("Start WriteTable on CLI for columns", zap.String("table", db.TableNames.Original))

	table := cli.newTableWithStream()

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

	tableWithFails := cli.newTableWithStream()
	tableWithFails.Header(tableHeadersForFailedColumns...)

	for _, failedColumn := range db.FailedParseColumns {
		if err := tableWithFails.Append([]any{
			failedColumn.OriginalName,
			failedColumn.CamelCaseName,
			failedColumn.LineNumber,
			failedColumn.Reason}); err != nil {
			cli.logger.Error("Failed to write table for failed parsed columns",
				zap.String("table", db.TableNames.Original),
				zap.Error(err))

			return err
		}
	}

	if len(db.FailedParseColumns) > 0 {
		cli.logger.Info("There were failed parsed columns for table",
			zap.String("table", db.TableNames.Original),
			zap.Int("failed_count", len(db.FailedParseColumns)))

		if err := tableWithFails.Render(); err != nil {
			cli.logger.Error("Failed to render table for failed parsed columns",
				zap.String("table", db.TableNames.Original),
				zap.Error(err))

			return err
		}
	}

	return nil
}

func (cli *TableWriterOnCLI) setTable() *tablewriter.Table {
	return tablewriter.NewTable(os.Stdout, tablewriter.WithRenderer(renderer.NewColorized(cli.colorCfg)),
		tablewriter.WithConfig(tablewriter.Config{
			Row: tw.CellConfig{
				Formatting:   tw.CellFormatting{AutoWrap: tw.WrapNormal}, // Wrap long content
				Alignment:    tw.CellAlignment{Global: tw.AlignLeft},     // Left-align rows
				ColMaxWidths: tw.CellWidth{Global: 25},
			},
			Footer: tw.CellConfig{
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
		}))
}

func (cli *TableWriterOnCLI) newTableWithStream() *tablewriter.Table {
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
		tablewriter.WithStreaming(tw.StreamConfig{
			Enable: true,
		}),
	)
}
