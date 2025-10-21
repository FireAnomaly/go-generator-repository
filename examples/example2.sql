CREATE TABLE IF NOT EXISTS testTable2
(
    id INT AUTO_INCREMENT PRIMARY KEY,
    TestText TEXT NOT NULL,
    TestInt INT DEFAULT 34534,
    TestBool BOOL,
    TestBoolean BOOLEAN,
    TestBoolButTinyInt TINYINT(1) DEFAULT 0,
    TestDate DATE,
    TestUnique TEXT UNIQUE,
    TestForeign INT
);