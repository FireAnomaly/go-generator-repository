package cli

import (
	"os"
	"os/exec"
	"runtime"
)

type writer struct {
	MaxRows      int
	MinRows      int
	DisabledRows map[int]bool
	SelectedRow  int
}

const minRows = 1 // really? :)

func newWriter(selectedRow int, countRows int) *writer {
	w := &writer{
		MaxRows:      countRows,
		MinRows:      minRows,
		SelectedRow:  selectedRow,
		DisabledRows: make(map[int]bool),
	}

	return w
}

func (w *writer) upRow() {
	selectedRow := w.SelectedRow
	selectedRow--
	if w.MinRows > selectedRow {
		return
	}

	w.SelectedRow = selectedRow
}

func (w *writer) downRow() {
	selectedRow := w.SelectedRow
	selectedRow++
	if w.MaxRows < selectedRow {
		return
	}

	w.SelectedRow = selectedRow
}

func (w *writer) addDeletedRow() {
	w.DisabledRows[w.SelectedRow] = true
	return
}

func (w *writer) restoreRow() {
	delete(w.DisabledRows, w.SelectedRow)
	return
}

func clearConsole() {
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
