package cli

import (
	"fmt"
	"strconv"

	"generatorFromMigrations/model"
	"github.com/gdamore/tcell/v2"
	"github.com/rivo/tview"
)

type Tview struct{}

func NewTview() *Tview {
	return &Tview{}
}

func (cli *Tview) ManageTableByUser(dbs []model.Database) error {
	cli.printDefault(dbs)

	return nil
}

func (cli *Tview) printDefault(dbs []model.Database) {
	app := tview.NewApplication()
	table := tview.NewTable().
		SetBorders(true)
	_, rows := 4, len(dbs)

	color := tcell.ColorWhite

	color = tcell.ColorYellow
	table.SetCell(0, 0,
		tview.NewTableCell("Number").
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

	table.SetCell(0, 1,
		tview.NewTableCell("Original Name").
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

	table.SetCell(0, 2,
		tview.NewTableCell("CamelCased Name").
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

	table.SetCell(0, 3,
		tview.NewTableCell("Columns Count").
			SetTextColor(color).
			SetAlign(tview.AlignCenter))

	for r := 0; r < rows; r++ {
		selectedColumn := 0

		table.SetCell(r+1, selectedColumn,
			tview.NewTableCell(strconv.Itoa(r)).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
		selectedColumn++

		table.SetCell(r+1, selectedColumn,
			tview.NewTableCell(dbs[r].TableNames.Original).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
		selectedColumn++

		table.SetCell(r+1, selectedColumn,
			tview.NewTableCell(dbs[r].TableNames.CamelCase).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
		selectedColumn++

		table.SetCell(r+1, selectedColumn,
			tview.NewTableCell(strconv.Itoa(len(dbs[r].Columns))).
				SetTextColor(color).
				SetAlign(tview.AlignCenter))
	}

	choisedDBs := make(map[int]string)

	table.Select(0, 0).SetFixed(1, 1).SetDoneFunc(func(key tcell.Key) {
		if key == tcell.KeyEscape {
			app.Stop()
		}
		if key == tcell.KeyEnter {
			table.SetSelectable(true, false)
		}
	}).SetSelectedFunc(func(row int, column int) {
		table.GetCell(row, column).SetTextColor(tcell.ColorRed)
		choisedDBs[row] = dbs[row-1].TableNames.Original
	})
	if err := app.SetRoot(table, true).SetFocus(table).Run(); err != nil {
		panic(err)
	}

	fmt.Println(choisedDBs)

}
