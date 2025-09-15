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

func NewRepository(logger *zap.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}

// Для маппинга

type Pair struct {
	Name    string
	Pointer any
	Scanner func()
}

type PairSlice []Pair

func (p PairSlice) Strings() []string {
	s := make([]string, len(p))
	for i, v := range p {
		s[i] = v.Name
	}

	return s
}

type ExampleModel struct {
	ID  NumID  `db:"ID"`
	Bla string `db:"Bla"`
}

/*
Где то в промежутке возвращать Pair и у него реализовать String и Addresses.
*/

// todo добавить фильтр для НЕ включать в SELECT

// SelectBuilder принимает поля из структуры. Или указывать все поля явно ИЛИ выбирать все поля, если поля не выбраны.
func (m *ExampleModel) SelectBuilder(fields ...any) (*sqlbuilder.SelectBuilder, error) {
	sb := sqlbuilder.NewSelectBuilder()

	parsed, err := m.parse(fields)
	if err != nil {
		return nil, err
	}

	sb.Select(parsed.Strings()...)
	/*
		Вытаскивать поля и пихать их в select
	*/

	return sb, nil
}

func (m *ExampleModel) Ptrs() []any {

}

func (m *ExampleModel) parse(args ...any) (PairSlice, error) {
	if m == nil {
		return nil, fmt.Errorf("model is nil")
	}

	rv := reflect.ValueOf(m).Elem()
	if rv.Kind() != reflect.Struct {
		return nil, fmt.Errorf("model is not a struct")
	}

	rt := rv.Type()

	type index struct {
		val        reflect.Value
		columnName string
	}

	byAddr := make(map[uintptr]index, rt.NumField())
	allIndex := make([]index, 0, rt.NumField())

	for i := 0; i < rt.NumField(); i++ {
		sf := rt.Field(i)
		if sf.PkgPath != "" {
			continue
		}

		columnName := sf.Tag.Get("db") // Поля без тега db не идут в запрос
		if columnName == "" {
			continue
		}

		fv := rv.Field(i)
		byAddr[fv.Addr().Pointer()] = index{
			val:        fv,
			columnName: columnName,
		}

		allIndex = append(allIndex, index{
			val:        fv,
			columnName: columnName,
		})
	}

	chosen := make([]index, 0, len(args))
	// fixme если поля не указаны, то ничего и не берётся - может стоить тогда брать все поля?
	for i, a := range args {
		av := reflect.ValueOf(a)
		if av.Kind() != reflect.Ptr || av.IsNil() {
			return nil, fmt.Errorf("arg %d must be non-nil pointer to a field of the same model", i)
		}
		mt, ok := byAddr[av.Pointer()]
		if !ok {
			return nil, fmt.Errorf("arg %d does not match any field of provided model", i)
		}
		chosen = append(chosen, mt)
	}

	out := make(PairSlice, 0, len(chosen))
	for _, c := range chosen {
		out = append(out, Pair{
			Name:    c.columnName,
			Pointer: c.val.Pointer(),
		})
	}

	return out, nil
}

func dereferencedValue(t reflect.Value) reflect.Value {
	for k := t.Kind(); k == reflect.Ptr || k == reflect.Interface; k = t.Kind() {
		t = t.Elem()
	}

	return t
}

var testStruct = new(ExampleModel)

var userStruct = sqlbuilder.NewStruct(new(ExampleModel))

// todo Добавить пример генератора шаблона
func (r *Repository) HAHAHA(ctx context.Context) error {
	ex := ExampleModel{}
	sb, err := ex.SelectBuilder(ex.ID, ex.Bla)
	if err != nil {
		return err
	}

	query, args := sb.Build()

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
