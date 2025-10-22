package cli

import (
	"fmt"
	"os"
	"time"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
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

	userChoice := make(chan string)

	go func() {
		cli.UserGetDB(dbs, userChoice)
	}()

	choice := <-userChoice

	println(choice)
	panic("stop")

}

func (cli *TableWriterOnCLI) UserGetDB(dbs []model.Database, userInput chan string) {
	select {
	case <-time.After(30 * time.Second):
		cli.logger.Warn("UserGetDB timeout")

		return

	default:
		cli.printDBs(dbs)

		var input string
		cli.logger.Info("Please enter the original name of the database you want to view columns for:")
		_, _ = fmt.Scanln(&input)

		userInput <- input
	}
}

func (cli *TableWriterOnCLI) calculateMoves() {
	cli.logger.Debug("Start CalculateMoves on CLI")
	keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.CtrlC, keys.Escape:
			return true, nil // Return true to stop listener
		case keys.Up:
			fmt.Println("\rYou pressed the Up arrow key")
		case keys.Down:
			fmt.Println("\rYou pressed the Down arrow key")
		case keys.Left:
			fmt.Println("\rYou pressed the Left arrow key")
		case keys.Right:
			fmt.Println("\rYou pressed the Right arrow key")
		// case keys.RuneKey: // Check if key is a rune key (a, b, c, 1, 2, 3, ...)
		// 	if key.String() == "q" { // Check if key is "q"
		// 		fmt.Println("\rQuitting application")
		// 		os.Exit(0) // Exit application
		// 	}
		//
		// 	fmt.Printf("\rYou pressed the rune key: %s\n", key)
		default:
			fmt.Printf("\rYou pressed: %s\n", key)
		}

		return false, nil // Return false to continue listening
	})
}

func (cli *TableWriterOnCLI) printDBs(dbs []model.Database) {
	cli.logger.Debug("Start WriteTable on CLI for databases")

	table := cli.newTable()

	table.Header([]any{"Num", "Original Name", "CamelCased Name", "Number of Columns"})

	num := 1
	for _, v := range dbs {
		if err := table.Append([]any{
			num,
			v.TableNames.Original,
			v.TableNames.CamelCase,
			len(v.Columns),
		}); err != nil {
			cli.logger.Error("Failed to write table database", zap.Error(err))
		}

		num++
	}

	if err := table.Render(); err != nil {
		cli.logger.Error("Failed to render table", zap.Error(err))
	}

	return
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

	tableWithFails := cli.newTable()
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
