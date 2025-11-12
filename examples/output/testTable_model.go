package output

import "time"

type TestTable struct {
	Id                 int               `db:"id"`
	TestText           string            `db:"TestText"`
	TestInt            int               `db:"TestInt"`
	TestBool           bool              `db:"TestBool"`
	TestBoolean        bool              `db:"TestBoolean"`
	TestBoolButTinyInt int               `db:"TestBoolButTinyInt"`
	TestDate           time.Time         `db:"TestDate"`
	TestUnique         string            `db:"TestUnique"`
	TestForeign        int               `db:"TestForeign"`
	TestJSON           []byte            `db:"TestJSON"`
	TestEnum           TestTableTestEnum `db:"TestEnum"`
}
type TestTableTestEnum string

const (
	TestTableTestEnumValue1 TestTableTestEnum = "Value1"
	TestTableTestEnumValue2 TestTableTestEnum = "Value2"
	TestTableTestEnumValue3 TestTableTestEnum = "Value3"
)
