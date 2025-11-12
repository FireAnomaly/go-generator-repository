package cli

import (
	"fmt"
	"os"
	"strings"

	"atomicgo.dev/keyboard"
	"atomicgo.dev/keyboard/keys"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/jedib0t/go-pretty/v6/text"
	"go.uber.org/zap"

	"github.com/FireAnomaly/go-generator-repository/model"
)

type columnWriter struct {
	writer     *writer
	userAction UserAction
	db         *model.Database
	logger     *zap.Logger
}

func newColumnWriter(db *model.Database, logger *zap.Logger) *columnWriter {
	if logger == nil {
		logger = zap.NewNop()
	}

	return &columnWriter{
		writer: newWriter(minRows, len(db.Columns)),
		db:     db,
		logger: logger,
	}
}

func (cw *columnWriter) manageColumns() (UserAction, error) {
	cw.writeColumns()

	if err := keyboard.Listen(cw.keyboardListenWrapperManageColumns); err != nil {
		return UserImmediatelyClose, err
	}

	return cw.userAction, nil

}

func (cw *columnWriter) keyboardListenWrapperManageColumns(key keys.Key) (stop bool, err error) {
	switch key.Code {
	case keys.Down:
		cw.writer.downRow()
		cw.writeColumns()

		return false, nil

	case keys.Up:
		cw.writer.upRow()
		cw.writeColumns()

		return false, nil

	case keys.Left:
		cw.userAction = UserWantBackToMainMenu
		return true, nil

	case keys.Enter:
		cw.userAction = UserAcceptChanges

		return true, nil

	case keys.CtrlC:
		cw.userAction = UserImmediatelyClose

		return true, nil

	default:
		cw.writeColumns()
		return false, nil
	}
}

func (cw *columnWriter) writeColumns() {
	cw.logger.Debug("Start writeColumns")

	clearConsole()

	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	t.AppendHeader(table.Row{"Original Name", "CamelCased Name", "Type", "DefaultValue", "EnumValues", "IsNull"})

	for _, column := range cw.db.Columns {
		t.AppendRow(table.Row{
			column.OriginalName,
			column.CamelCaseName,
			column.Type,
			column.DefaultValue,
			strings.Join(column.EnumValues, ", "),
			column.IsNull,
		})
		t.AppendSeparator()
	}

	t.SetRowPainter(cw.getRowPainter)

	t.Render()

	fmt.Printf("Left Arrow - to back on main screen \nCTRL+C - exit \nEnter - Apply\n>>>")

	return
}

func (cw *columnWriter) getRowPainter(row table.Row, attr table.RowAttributes) text.Colors {
	if attr.Number == cw.writer.SelectedRow {
		return text.Colors{text.FgGreen}
	}

	return text.Colors{text.FgWhite}
}
