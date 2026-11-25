# Go Tool Dependencies

> Начиная с Go 1.24, зависимости инструментов разработки (tool dependencies) могут управляться непосредственно через `go.mod` с использованием директивы `tool`.

Это позволяет фиксировать версии инструментов на уровне проекта, запускать их без глобальной установки и обеспечивать воспроизводимость окружения разработки.

## Add Tool Dependencies

```bash
go get -tool <package>@<version>
```

Например:

```bash
go get -tool github.com/air-verse/air@latest
go get -tool github.com/pressly/goose/v3/cmd/goose@latest
```

В `go.mod` автоматически добавляются директива `tool` и необходимые зависимости:

```go
tool (
    github.com/air-verse/air
    github.com/pressly/goose/v3/cmd/goose
)
```

## List Tool Dependencies

Просмотр доступных инструментов:

```bash
go tool
```

## Run Tool

Запуск инструмента по имени:

```bash
go tool <name> [arguments]
```

Например:

```bash
go tool air
go tool goose status
```

## Update Tool Dependencies

Обновление отдельного инструмента:

```bash
go get -tool <package>@latest
```

Обновление всех зарегистрированных инструментов:

```bash
go get tool@latest
```

## Remove Tool Dependencies

```bash
go get -tool <package>@none
go mod tidy
```

## Tool Dependencies vs Global Installation

| Command | Description |
|---|---|
| `go get -tool` | Регистрирует инструмент в `go.mod` и фиксирует его версию на уровне проекта. |
| `go tool` | Запускает зарегистрированный инструмент без глобальной установки. |
| `go install <package>@<version>` | Собирает и устанавливает исполняемый файл в `GOBIN` или `GOPATH/bin`, не изменяя зависимости текущего модуля. |

Для инструментов, используемых в процессах разработки, тестирования, генерации кода, миграции БД и CI/CD, предпочтительно использовать tool dependencies.

## References

- [Go Modules Reference — tool directive](https://go.dev/ref/mod#go-mod-file-tool)
- [Go 1.24 Release Notes — Tools](https://go.dev/doc/go1.24#tools)