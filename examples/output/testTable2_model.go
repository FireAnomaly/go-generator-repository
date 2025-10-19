package testTable2_model

type TestTable2 struct {
    Id int `db:"id"`
    TestText string `db:"TestText"`
    TestInt int `db:"TestInt"`
    TestBool bool `db:"TestBool"`
    TestBoolean bool `db:"TestBoolean"`
    TestBoolButTinyInt int `db:"TestBoolButTinyInt"`
    TestDate time.Time `db:"TestDate"`
    TestUnique string `db:"TestUnique"`
    TestForeign int `db:"TestForeign"`
    TestEnum string `db:"TestEnum"`
}
