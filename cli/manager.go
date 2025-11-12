package cli

import (
	"errors"

	"go.uber.org/zap"

	"github.com/FireAnomaly/go-generator-repository/model"
)

type UserAction int

const (
	UserImmediatelyClose UserAction = iota
	UserWantDiveToColumns
	UserAcceptChanges
	UserWantBackToMainMenu
)

type TableWriterOnCLI struct {
	table  *tableWriter
	logger *zap.Logger
}

func NewTableWriterOnCLI(logger *zap.Logger, dbs []*model.Database) *TableWriterOnCLI {
	if logger == nil {
		logger = zap.NewNop()
	}

	cli := &TableWriterOnCLI{
		table:  newTableWriter(logger, dbs),
		logger: logger.Named("CLI Table Writer: "),
	}

	return cli
}

func (cli *TableWriterOnCLI) ManageTableByUser() error {
	cli.logger.Debug("Start ManageTableByUser on CLI")

	return cli.manageWriters()
}

var CloseSignal = errors.New("user initiate close signal")

func (cli *TableWriterOnCLI) manageWriters() error {
	for {
		userAction, chosenDB, err := cli.table.manageDBs()
		if err != nil {
			return err
		}

		columnwriter := newColumnWriter(chosenDB, cli.logger)
	choice:
		switch userAction {
		case UserImmediatelyClose:
			clearConsole()
			return CloseSignal
		case UserWantDiveToColumns:
			userAction, err = columnwriter.manageColumns()
			if err != nil {
				return err
			}

			goto choice
		case UserAcceptChanges:
			clearConsole()
			return nil
		case UserWantBackToMainMenu:
			break
		}
	}
}
