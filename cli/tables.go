package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/FireAnomaly/go-generator-repository/model"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/zap"
)

type TableWriterOnCLI struct {
	writer     *cliParams
	managedDBs *managedDBs
	dbs        []model.Database
	logger     *zap.Logger
}

func NewTableWriterOnCLI(logger *zap.Logger) *TableWriterOnCLI {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &TableWriterOnCLI{logger: logger.Named("CLI Table Writer: ")}
}

func (cli *TableWriterOnCLI) ManageTableByUser(dbs []model.Database) error {
	cli.logger.Debug("Start ManageTableByUser on CLI")

	return cli.manageWriters(dbs)
}

func (cli *TableWriterOnCLI) clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		cmd.Run()
	default:
		panic("Your platform is unsupported! I can't clearConsole terminal screen :(")
	}
}

func (cli *TableWriterOnCLI) manageWriters(dbs []model.Database) error {
	for {
		err := cli.manageDBs(dbs)
		if err != nil {
			return err
		}

		if cli.managedDBs.isWannaExit {
			return nil
		}

		if cli.managedDBs.isNeedToManageCols {
			if err = cli.choiceColumns(dbs[cli.writer.SelectedRow-1]); err != nil {
				return err
			}
		}
	}
}

type managedDBs struct {
	isWannaExit        bool
	isNeedToManageCols bool
	selectedRow        int
}

func (cli *TableWriterOnCLI) keyboardListenWrapperManageDBs(key keys.Key) (stop bool, err error) {
	switch key.Code {
	case keys.Down:
		cli.downRow()
		cli.writeTable(cli.dbs)

		return false, nil

	case keys.Up:
		cli.upRow()
		cli.writeTable(cli.dbs)

		return false, nil

	case keys.Right:
		cli.managedDBs = &managedDBs{
			isWannaExit:        false,
			isNeedToManageCols: true,
			selectedRow:        cli.writer.SelectedRow,
		}

		return true, nil

	case keys.CtrlC:
		cli.managedDBs = &managedDBs{
			isWannaExit: true,
		}
		return true, nil

	default:
		return false, nil

	}
}

func (cli *TableWriterOnCLI) manageDBs(dbs []model.Database) error {
	selectedRow := minRows
	rows := len(dbs)

	cli.initDescriptor(selectedRow, rows)
	cli.writeTable(dbs)

	err := keyboard.Listen(cli.keyboardListenWrapperManageDBs)
	if err != nil {
		return err
	}

	return nil
}

func (cli *TableWriterOnCLI) choiceColumns(db model.Database) error {
	selectedRow := minRows
	rows := len(db.Columns)

	cli.initDescriptor(selectedRow, rows)
	cli.writeTableColumns(db)

	return keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Down:
			cli.downRow()
			cli.writeTableColumns(db)

			return false, nil

		case keys.Up:
			cli.upRow()
			cli.writeTableColumns(db)

			return false, nil

		case keys.Left:
			return true, nil

		case keys.CtrlC:
			return true, nil

		default:
			return false, nil
		}
	})
}

type cliParams struct {
	MaxRows     int
	MinRows     int
	Message     *string
	SelectedRow int
}

const minRows = 1 // really? :)

func (cli *TableWriterOnCLI) initDescriptor(selectedRow int, countRows int) {
	cli.writer = &cliParams{
		MaxRows:     countRows,
		MinRows:     minRows,
		SelectedRow: selectedRow,
	}

	return
}

func (cli *TableWriterOnCLI) upRow() {
	selectedRow := cli.writer.SelectedRow
	selectedRow--
	if cli.writer.MinRows > selectedRow {
		return
	}

	cli.writer.SelectedRow = selectedRow
}

func (cli *TableWriterOnCLI) downRow() {
	selectedRow := cli.writer.SelectedRow
	selectedRow++
	if cli.writer.MaxRows < selectedRow {
		return
	}

	cli.writer.SelectedRow = selectedRow
}

func (cli *TableWriterOnCLI) writeTable(dbs []model.Database) {
	cli.logger.Debug("Start writeTable")

	cli.clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Line Number"})

	currentRow := 1
	for _, db := range dbs {
		t.AppendRow(table.Row{db.TableNames.Original, db.TableNames.CamelCase, len(db.Columns)})
		t.AppendSeparator()

		t.SetRowPainter(func(row table.Row, attr table.RowAttributes) text.Colors {
			if attr.Number == cli.writer.SelectedRow {
				return text.Colors{text.FgGreen}
			}

			return text.Colors{text.FgWhite}
		})
		currentRow++
	}

	t.Render()

	return
}

func (cli *TableWriterOnCLI) writeTableColumns(db model.Database) {
	cli.logger.Debug("Start writeTableColumns")

	cli.clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Type", "DefaultValue", "EnumValues", "IsNull"})

	currentRow := 1
	for _, column := range db.Columns {
		t.AppendRow(table.Row{
			column.OriginalName,
			column.CamelCaseName,
			column.Type,
			column.DefaultValue,
			strings.Join(column.EnumValues, ", "),
			column.IsNull,
		})
		t.AppendSeparator()

		t.SetRowPainter(func(row table.Row, attr table.RowAttributes) text.Colors {
			if attr.Number == cli.writer.SelectedRow {
				return text.Colors{text.FgGreen}
			}

			return text.Colors{text.FgWhite}
		})
		currentRow++
	}

	t.Render()

	return
}
