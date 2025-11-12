package cli

import (
	"fmt"
	"os"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/zap"

	"github.com/FireAnomaly/go-generator-repository/model"
)

type tableWriter struct {
	logger     *zap.Logger
	dbs        []*model.Database
	userAction UserAction
	writer     *writer
}

func newTableWriter(logger *zap.Logger, dbs []*model.Database) *tableWriter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &tableWriter{
		logger: logger,
		dbs:    dbs,
		writer: newWriter(minRows, len(dbs)),
	}
}

func (tw *tableWriter) manageDBs() (UserAction, *model.Database, error) {
	tw.writeTable()

	err := keyboard.Listen(tw.keyboardListenWrapperManageDBs)
	if err != nil {
		return UserImmediatelyClose, nil, err
	}

	return tw.userAction, tw.dbs[tw.writer.SelectedRow-1], nil
}

func (tw *tableWriter) keyboardListenWrapperManageDBs(key keys.Key) (stop bool, err error) {
	if key.String() == "r" {
		tw.enableDB()
		tw.writeTable()
	}

	switch key.Code {
	case keys.Down:
		tw.writer.downRow()
		tw.writeTable()

		return false, nil

	case keys.Up:
		tw.writer.upRow()
		tw.writeTable()

		return false, nil

	case keys.Right:
		tw.userAction = UserWantDiveToColumns

		return true, nil

	case keys.Enter:
		tw.userAction = UserAcceptChanges

		return true, nil

	case keys.CtrlC:
		tw.userAction = UserImmediatelyClose
		return true, nil

	case keys.Backspace:
		tw.disableDB()

		tw.writeTable()

		return false, nil

	default:
		tw.writeTable()
		return false, nil
	}
}

func (tw *tableWriter) disableDB() {
	tw.dbs[tw.writer.SelectedRow-1].Disabled = true
	tw.writer.addDeletedRow()
}

func (tw *tableWriter) enableDB() {
	tw.dbs[tw.writer.SelectedRow-1].Disabled = false
	tw.writer.restoreRow()
}

func (tw *tableWriter) writeTable() {
	tw.logger.Debug("Start writeTable")

	clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Line Number"})

	for _, db := range tw.dbs {
		t.AppendRow(table.Row{db.TableNames.Original, db.TableNames.CamelCase, len(db.Columns)})
		t.AppendSeparator()
	}

	t.SetRowPainter(tw.getRowPainter)

	t.Render()

	fmt.Printf("Right Arrow - Dive to selected base \nCTRL+C - Exit \nEnter - Apply \nBackspace - Disable Database\nr - To Restore Disabled\n>>>")

	return
}

func (tw *tableWriter) getRowPainter(row table.Row, attr table.RowAttributes) text.Colors {
	if tw.writer.DisabledRows[attr.Number] {
		return text.Colors{text.FgRed}
	}

	if attr.Number == tw.writer.SelectedRow {
		return text.Colors{text.FgGreen}
	}

	return text.Colors{text.FgWhite}
}
