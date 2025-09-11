package main

import "errors"

var (
	ErrMigrationNotFound = errors.New("migration not found")
	ErrInvalidMigration  = errors.New("migration is not valid")
	ErrInvalidRegExp     = errors.New("invalid regexp")
)

type Database struct {
	TableName string
	Columns   []Column
}

type Column struct {
	Name   string
	Type   string
	IsNull bool
}
