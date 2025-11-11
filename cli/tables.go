package cli

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/zap"

	"github.com/FireAnomaly/go-generator-repository/model"
)

type TableWriterOnCLI struct {
	writer     *cliParams
	managedDBs *managedDBs
	selectedDB *model.Database
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
	cli.dbs = dbs

	return cli.manageWriters()
}

func (cli *TableWriterOnCLI) clearConsole() {
	switch runtime.GOOS {
	case "windows":
		cmd := exec.Command("cmd", "/c", "cls")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	case "linux", "darwin":
		cmd := exec.Command("clear")
		cmd.Stdout = os.Stdout
		_ = cmd.Run()
	default:
		panic("Your platform is unsupported! I can't clearConsole terminal screen :(")
	}
}

var CloseSignal = errors.New("user initiate close signal")

func (cli *TableWriterOnCLI) manageWriters() error {
	for {
		err := cli.manageDBs()
		if err != nil {
			return err
		}

		if cli.managedDBs.saveAndExit {
			return nil
		}

		if cli.managedDBs.exit {
			return CloseSignal
		}

		if cli.managedDBs.isNeedToManageCols {
			cli.selectedDB = &cli.dbs[cli.managedDBs.selectedRow-1]
			if err = cli.manageColumns(); err != nil {
				return err
			}

			if cli.managedDBs.exit {
				return CloseSignal
			}
		}
	}
}

type managedDBs struct {
	db                 *model.Database
	exit               bool
	disableCurrent     bool
	saveAndExit        bool
	isNeedToManageCols bool
	selectedRow        int
}

func (cli *TableWriterOnCLI) manageDBs() error {
	selectedRow := minRows
	rows := len(cli.dbs)

	cli.initDescriptor(selectedRow, rows)
	cli.writeTable()

	err := keyboard.Listen(cli.keyboardListenWrapperManageDBs)
	if err != nil {
		return err
	}

	return nil
}

func (cli *TableWriterOnCLI) keyboardListenWrapperManageDBs(key keys.Key) (stop bool, err error) {
	switch key.Code {
	case keys.Down:
		cli.downRow()
		cli.writeTable()

		return false, nil

	case keys.Up:
		cli.upRow()
		cli.writeTable()

		return false, nil

	case keys.Right:
		cli.managedDBs = &managedDBs{
			exit:               false,
			isNeedToManageCols: true,
			selectedRow:        cli.writer.SelectedRow,
		}

		return true, nil

	case keys.Enter:
		cli.managedDBs = &managedDBs{
			saveAndExit: true,
		}

		return true, nil

	case keys.CtrlC:
		cli.managedDBs = &managedDBs{
			exit: true,
		}
		return true, nil
	case keys.Backspace:
		cli.selectedDB.Disabled = true
		return true, nil

	default:
		return false, nil
	}
}

func (cli *TableWriterOnCLI) manageColumns() error {
	selectedRow := minRows
	rows := len(cli.selectedDB.Columns)

	cli.initDescriptor(selectedRow, rows)
	cli.writeColumns()

	return keyboard.Listen(cli.keyboardListenWrapperManageColumns)
}

func (cli *TableWriterOnCLI) keyboardListenWrapperManageColumns(key keys.Key) (stop bool, err error) {
	switch key.Code {
	case keys.Down:
		cli.downRow()
		cli.writeColumns()

		return false, nil

	case keys.Up:
		cli.upRow()
		cli.writeColumns()

		return false, nil

	case keys.Left:
		return true, nil

	case keys.Enter:
		return true, nil

	case keys.CtrlC:
		cli.managedDBs = &managedDBs{
			exit: true,
		}

		return true, nil

	default:
		return false, nil
	}
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

func (cli *TableWriterOnCLI) writeTable() {
	cli.logger.Debug("Start writeTable")

	cli.clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Line Number"})

	currentRow := 1
	for _, db := range cli.dbs {
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

	fmt.Printf("Right Arrow - to insert selected base \nCTRL+C - exit \nEnter - Apply \n>>>")

	return
}

func (cli *TableWriterOnCLI) writeColumns() {
	cli.logger.Debug("Start writeColumns")

	cli.clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Type", "DefaultValue", "EnumValues", "IsNull"})

	currentRow := 1
	for _, column := range cli.selectedDB.Columns {
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

	fmt.Printf("Left Arrow - to back on main screen \nCTRL+C - exit \nEnter - Apply\n>>>")

	return
}
