package main

import (
	"context"
	"database/sql"
	"fmt"
	"reflect"

	"github.com/huandu/go-sqlbuilder"
	"go.uber.org/zap"
)

// todo Прокинуть телеметрию и транзакцию
type Repository struct {
	logger *zap.Logger
	db     *sql.DB
}

// Для маппинга

type Pair struct {
	Name    string
	Pointer any
}

type PairSlice []*Pair

type ExampleModel struct {
	ID  NumID   `db:"ID"`
	Bla *string `db:"Bla"`
}

/*
Где то в промежутке возвращать Pair и у него реализовать String и Addresses.
*/

// SelectBuilder принимает поля из структуры
func (m *ExampleModel) SelectBuilder(fields ...any) *sqlbuilder.SelectBuilder {
	sb := sqlbuilder.NewSelectBuilder()

	parsed := m.parse(fields)
	/*
		Вытаскивать поля и пихать их в select
	*/
	return sb
}

func (m *ExampleModel) parse(s ...any) []Pair {
	var pair []Pair
	for k, v := range s {
		tRef := reflect.TypeOf(v)
		name := tRef.Name()

		valOf := reflect.ValueOf(v)
		refVal := dereferencedValue(valOf)

		addr := refVal.FieldByName(name).Addr().Interface()

	}
}

func dereferencedValue(t reflect.Value) reflect.Value {
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()
	}

	return t
}

var testStruct = new(ExampleModel)

func NewRepository(logger *zap.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}

var userStruct = sqlbuilder.NewStruct(new(ExampleModel))

// todo Добавить пример генератора шаблона
func (r *Repository) HAHAHA(ctx context.Context) error {
	testStruct.SelectBuilder(testStruct.ID, testStruct.Bla)

	sb := sqlbuilder.NewSelectBuilder()
	sb.Select(testStruct.Set(testStruct.Bla, testStruct.Bla)...)

	println(query)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return err
	}

	var models []ExampleModel
	for rows.Next() {
		var model ExampleModel
		err = rows.Scan(userStruct.Addr(&model)...) // userStruct.Addr(&model)... -> возвращает слайс указателей на поля структуры
		if err != nil {
			return err
		}

		models = append(models, model)
	}

	for _, v := range models {
		fmt.Println(v.ID)
	}

	return nil
}
