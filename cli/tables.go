package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/FireAnomaly/go-generator-repository/model"
	"github.com/FireAnomaly/go-keyboard-capture"
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

func (cli *TableWriterOnCLI) UserGetDB(dbs []model.Database) {
	events := make(chan keyboard.KeyEvent)
	userStop := make(chan bool)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	go func() {
		cli.captureKeyboard(ctx, events, userStop)
	}()

	table := cli.newTable()
	cli.streamDBsTable(table, dbs) // Initial render
	// defer func(table *tablewriter.Table) {
	// 	err := table.Close()
	// 	if err != nil {
	// 		cli.logger.Error("Failed to close table", zap.Error(err))
	// 	}
	// }(table)
	//
	// if err := table.Start(); err != nil {
	// 	cli.logger.Fatal("Failed to start table streaming", zap.Error(err))
	// }

	for {
		select {
		case <-userStop:
			cli.logger.Info("User requested to stop table management")
			cancel()

			return
		case <-events:
			cli.streamDBsTable(table, dbs)
		}
	}
}

func (cli *TableWriterOnCLI) captureKeyboard(ctx context.Context, events chan keyboard.KeyEvent, userStop chan bool) {
	cli.logger.Debug("Start captureKeyboard on CLI")

	err := keyboard.CaptureKeyboard(ctx, events)
	if err != nil {
		cli.logger.Error("Failed to capture keyboard", zap.Error(err))
		panic(err)
	}

	for event := range events {
		if event.Key == 'q' {
			cli.logger.Info("Quit keyboard captured")
			userStop <- true

			break
		}

		if event.Key == 'c' {
			events <- event
		}

		if event.Code == keyboard.KeyUpArrow {
			events <- event
			// fmt.Println("Up Arrow Pressed")
		}

		if event.Code == keyboard.KeyDownArrow {
			events <- event
			// fmt.Println("Down Arrow Pressed")
		}

	}

	return
}

func (cli *TableWriterOnCLI) streamDBsTable(table *tablewriter.Table, dbs []model.Database) {
	cli.logger.Debug("Start WriteTable on CLI for databases")

	// println("")

	fmt.Print("\r")

	table.Header([]string{"Num", "Original Name", "CamelCased Name", "Number of Columns"})

	num := 1
	for _, v := range dbs {
		if err := table.Append([]string{
			strconv.Itoa(num),
			v.TableNames.Original,
			v.TableNames.CamelCase,
			strconv.Itoa(len(v.Columns)),
		}); err != nil {
			cli.logger.Error("Failed to write table database", zap.Error(err))
		}

		num++
	}

	table.Render()

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
				Alignment: tw.CellAlignment{Global: tw.AlignCenter},
			},
		}),
	)
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
