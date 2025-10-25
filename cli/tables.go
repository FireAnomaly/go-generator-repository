package cli

import (
	"os"
	"os/exec"
	"runtime"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/FireAnomaly/go-generator-repository/model"
	"github.com/aws/smithy-go/ptr"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/zap"
)

type TableWriterOnCLI struct {
	logger *zap.Logger
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

func (cli *TableWriterOnCLI) clearConsole() {
	value, ok := clearConsole[runtime.GOOS] // runtime.GOOS -> linux, windows, darwin etc.
	if ok {                                 // if we defined a clearConsole func for that platyform:
		value() // we execute it
	} else { // unsupported platform
		panic("Your platform is unsupported! I can't clearConsole terminal screen :(")
	}
}

func (cli *TableWriterOnCLI) manageWriters(dbs []model.Database) error {
	for {
		chosenDB, err := cli.choiceDataBase(dbs)
		if err != nil {
			return err
		}

		if chosenDB.isWannaExit {
			return nil
		}

		if chosenDB.isNeedToManageCols {
			if err = cli.choiceColumns(dbs[chosenDB.selectedRow-1]); err != nil {
				return err
			}
		}
	}
}

type resultOfChoisenDatabase struct {
	isWannaExit        bool
	isNeedToManageCols bool
	selectedRow        int
}

func (cli *TableWriterOnCLI) choiceDataBase(dbs []model.Database) (resultOfChoisenDatabase, error) {
	var result resultOfChoisenDatabase

	selectedRow := minRows
	rows := len(dbs)

	writeDescriptor := cli.writeDescriptor(selectedRow, rows)
	cli.writeTable(writeDescriptor.SelectedRow, dbs)

	err := keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Down:
			writeDescriptor = cli.writeDescriptor(writeDescriptor.SelectedRow+1, len(dbs))
			cli.writeTable(writeDescriptor.SelectedRow, dbs)

			return false, nil

		case keys.Up:
			writeDescriptor = cli.writeDescriptor(writeDescriptor.SelectedRow-1, len(dbs))
			cli.writeTable(writeDescriptor.SelectedRow, dbs)

			return false, nil

		case keys.Right:
			result.isNeedToManageCols = true
			result.selectedRow = selectedRow

			return true, nil

		case keys.CtrlC:
			result.isWannaExit = true

			return true, nil

		default:
			return false, nil
		}
	})
	if err != nil {
		return result, err
	}

	return result, nil
}

func (cli *TableWriterOnCLI) choiceColumns(db model.Database) error {
	selectedRow := minRows
	rows := len(db.Columns)

	writeDescriptor := cli.writeDescriptor(selectedRow, rows)
	cli.writeTableColumns(writeDescriptor.SelectedRow, db)

	return keyboard.Listen(func(key keys.Key) (stop bool, err error) {
		switch key.Code {
		case keys.Down:
			writeDescriptor = cli.writeDescriptor(writeDescriptor.SelectedRow+1, rows)
			cli.writeTableColumns(writeDescriptor.SelectedRow, db)

			return false, nil

		case keys.Up:
			writeDescriptor = cli.writeDescriptor(writeDescriptor.SelectedRow-1, rows)
			cli.writeTableColumns(writeDescriptor.SelectedRow, db)

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

type toWrite struct {
	MaxRows     int
	MinRows     int
	Message     *string
	SelectedRow int
}

const minRows = 1 // really? :)

func (cli *TableWriterOnCLI) writeDescriptor(selectedRow int, countRows int) toWrite {
	writer := toWrite{
		MaxRows:     countRows,
		MinRows:     minRows,
		Message:     nil,
		SelectedRow: selectedRow,
	}

	if selectedRow < minRows {
		writer.Message = ptr.String("Already at the top row")
		writer.SelectedRow = minRows
	}

	if selectedRow > countRows {
		writer.Message = ptr.String("Already at the bottom row")
		writer.SelectedRow = countRows
	}

	return writer
}

func (cli *TableWriterOnCLI) writeTable(selectedRow int, dbs []model.Database) {
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
			if attr.Number == selectedRow {
				return text.Colors{text.FgGreen}
			}

			return text.Colors{text.FgWhite}
		})
		currentRow++
	}

	t.Render()

	return
}

func (cli *TableWriterOnCLI) writeTableColumns(selectedRow int, db model.Database) {
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
			if attr.Number == selectedRow {
				return text.Colors{text.FgGreen}
			}

			return text.Colors{text.FgWhite}
		})
		currentRow++
	}

	t.Render()

	return
}
