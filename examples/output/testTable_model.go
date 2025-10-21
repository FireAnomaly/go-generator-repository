package output 
 
import "time" 
 
type TestTable struct {
    Id int `db:"id"`
    TestText string `db:"TestText"`
    TestInt int `db:"TestInt"`
    TestBool bool `db:"TestBool"`
    TestBoolean bool `db:"TestBoolean"`
    TestBoolButTinyInt int `db:"TestBoolButTinyInt"`
    TestDate time.Time `db:"TestDate"`
    TestUnique string `db:"TestUnique"`
    TestForeign int `db:"TestForeign"`
    TestJSON []byte `db:"TestJSON"`
    TestEnum TestEnum `db:"TestEnum"`
} r

type TestEnum string 

const (
    TestEnumValue1 TestEnum = "Value1"
    TestEnumValue2 TestEnum = "Value2"
    TestEnumValue3 TestEnum = "Value3"
)
 
