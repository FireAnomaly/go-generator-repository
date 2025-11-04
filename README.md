# go-generator-repository

A CLI tool for generating Go models from MySQL migration files.

## Features

- Parses MySQL migration files to extract database schema.
- Generates Go structs and custom types (enums) based on the schema.
- Interactive CLI mode for selecting tables to generate models for.
- Configurable logging with levels.

## Requirements

- Go 1.25+
- MySQL migration files

## Installation

Clone the repository and build:

```sh
git clone https://github.com/FireAnomaly/go-generator-repository.git
cd go-generator-repository
go build -o go-generator-repo .
```

Or install directly:

```sh
go install github.com/FireAnomaly/go-generator-repository@latest
```

## Usage

Run in the working directory of your project:

```sh
./go-generator-repo -in path/to/migrations -out path/to/output [flags]
```

The tool will interactively prompt you to select databases and tables for model generation.

### Flags

- `-in` (required): Path to the directory containing migration files.
- `-out` (required): Path to save generated models.
- `-log`: Enable detailed logging (optional).
- `-loglevel`: Set the logging level (optional, default: info). Options: debug, info, warn, error, fatal, panic.

### Example

```sh
./go-generator-repo -in ./migrations -out ./models -log -loglevel debug
```

## Project Structure

- `main.go`: Entry point, CLI parsing, and workflow orchestration.
- `cli/`: Interactive CLI utilities for table selection.
- `model/`: Data structures for database schema representation.
- `parsers/mysql/`: MySQL migration file parser.
- `templater/`: Go code generation templates and logic.
- `examples/`: Sample MySQL migration files.

## Logging

Uses [uber-go/zap](https://github.com/uber-go/zap) for logging. Enable with `-log` flag and set level with `-loglevel`.

## Generated Output

For each selected table, generates a Go file with:
- A struct representing the table.
- Custom types for enum columns.
- Constants for enum values.

Example generated code:

```go
package models

type TestTable struct {
    ID                int    `json:"id" db:"id"`
    TestText          string `json:"test_text" db:"TestText"`
    TestInt           int    `json:"test_int" db:"TestInt"`
    TestBool          bool   `json:"test_bool" db:"TestBool"`
    TestDate          string `json:"test_date" db:"TestDate"`
    TestUnique        string `json:"test_unique" db:"TestUnique"`
    TestForeign       int    `json:"test_foreign" db:"TestForeign"`
    TestJSON          string `json:"test_json" db:"TestJSON"`
    TestEnum          TestEnum `json:"test_enum" db:"TestEnum"`
}

type TestEnum string

const (
    TestEnumValue1 TestEnum = "Value1"
    TestEnumValue2 TestEnum = "Value2"
    TestEnumValue3 TestEnum = "Value3"
)
```

## ToDos

- [ ] Improve graphic interface (currently only for table selection).
- [ ] Upgrade templater to support complex relationships (foreign keys, many-to-many, etc.).
- [ ] Upgrade parser to support more SQL dialects.
- [ ] Add support for generating relationships between models (e.g., ToModel() and FromModel() methods).

## License

MIT
