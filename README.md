# go-generator-repository

A CLI tool for generating Go models from MySQL migration files.

## Features

- Parses MySQL migration files to extract database schema.
- Generates Go structs and enums based on the schema.
- Supports CLI and interactive (graphic) mode.
- Configurable logging.

## Requirements

- Go 1.18+
- MySQL migration files

## Installation

Clone the repository and build:

```sh
go install https://github.com/FireAnomaly/go-generator-repository@latest
```

## Usage

```sh
./go-generator-repo -in path/to/migrations -out path/to/output [flags]
```

Run in workdirectory of your project

### Flags

- `-in` (required): Path to the migration files.
- `-out` (required): Path to save generated models.
- `-graphic`: Enable ~~interactive~~ table selection (optional). (true or false) Ctrl + C to exit.
- `-log`: Enable detailed logging (optional). (true or false)

### Example

```sh
./go-generator-repo -in ./migrations -out ./models -log
```

## Project Structure

- `main.go`: Entry point, CLI parsing, and workflow orchestration.
- `cli/`: CLI utilities and interactive mode.
- `model/`: Data structures for database schema.
- `parsers/mysql/`: MySQL migration parser.
- `templater/`: Go code generation templates.

## Logging

Enable detailed logs with `-log` flag. Uses [uber-go/zap](https://github.com/uber-go/zap).

## ToDos

-[ ] Get functional to graphic interface (now is only to show)

-[ ] Upgrade templater to support more complex relationships
- Like to other tables (foreign keys), many to many, etc.
  
-[ ] Upgrade parser to support more SQL dialects

-[ ] Upgrade custom types that templater can generate
- Add relation with other models in your project, like ToModel() and FromModel()


## License

MIT


---

