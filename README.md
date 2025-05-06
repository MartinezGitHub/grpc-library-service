# Library Service

Проект написан на языке go.

## Технологии
* [zap](https://github.com/uber-go/zap) для логирования.
* [gRPC gateway](https://grpc-ecosystem.github.io/grpc-gateway/) для поддержки REST-to-gRPC API.
* [protobuf](https://protobuf.dev/)
* [mockgen](https://github.com/golang/mock) для тестирования верхних уровней.
* [PostgreSQL](https://www.postgresql.org/) используемая база данных.

## Архитектура
Упрощенная версия [go-clean-template](https://github.com/evrone/go-clean-template).

## Сервис Library

### Книги
| RPC Метод       | HTTP Эндпоинт              | Тип    | Описание                          | Request                 | Response                     |
|-----------------|---------------------------|--------|----------------------------------|-------------------------|-----------------------------|
| `AddBook`       | POST /v1/library/book     | Unary  | Добавление новой книги           | `name`, `author_ids[]`  | `Book` объект               |
| `UpdateBook`    | PUT /v1/library/book      | Unary  | Обновление книги                 | `id`, `name`, `author_ids[]` | Пустой ответ                |
| `GetBookInfo`   | GET /v1/library/book/{id} | Unary  | Получение информации о книге     | `id`                    | `Book` объект               |

### Авторы
| RPC Метод           | HTTP Эндпоинт                   | Тип      | Описание                          | Request                         | Response                     |
|---------------------|--------------------------------|----------|----------------------------------|---------------------------------|-----------------------------|
| `RegisterAuthor`    | POST /v1/library/author        | Unary    | Регистрация нового автора        | `name` | `id`                |
| `ChangeAuthorInfo`  | PUT /v1/library/author         | Unary    | Изменение информации об авторе    | `id`, `name`                    | Пустой ответ                |
| `GetAuthorInfo`     | GET /v1/library/author/{id}    | Unary    | Получение информации об авторе    | `id`                      | `id`, `name`                |
| `GetAuthorBooks`    | GET /v1/library/author_books/{author_id} | Server Streaming | Получение книг автора | `author_id`| Поток `Book` объектов |

## Структуры данных

### Book
```protobuf
message Book {
  string id = 1;                   
  string name = 2;
  repeated string author_id = 3;     
  google.protobuf.Timestamp created_at = 4;
  google.protobuf.Timestamp updated_at = 5;
}
```

## Outbox
Также в этом сервисе реализован паттерн outbox для работы с запросами в сторонние сервисы.
### Файлы
* [outbox.go (repository)](internal/usecase/repository/outbox.go)
* [outbox.go (outbox use case)](internal/usecase/outbox/outbox.go)

## Сборка
`make generate`

`make build`

После этого можно запускать тесты.

## Тестирование
Тестирование разбивается на уровни:
* [library_test.go](internal/usecase/library/library_test.go) тесты слоя UseCase, где реализована вся бизнес-логика сервиса.
* [controller_test.go](internal/controller/controller_test.go) тесты транспортного уровня/слоя валидации.
* [integration_test.go](integration-test/outbox_hw/integration_test.go) интеграционные тесты.

Отдельно тестируются:
* [config_test.go](config/config.go) конфигурация.
* [kind_handler_error_test.go](internal/errors/kind_handler_error_test.go).








[//]: # ()
[//]: # ()
[//]: # ()
[//]: # (# HW 1 &#40;database&#41;)

[//]: # ()
[//]: # (В этом домашнем задании вам предстоит реализовать интеграцию с базой данных в рамках сервиса **library**.)

[//]: # (Для простоты понимания описание этого ДЗ сделано в императивном, а не декларативном стиле)

[//]: # ()
[//]: # (## Часть 1)

[//]: # (Ниже описана одна из возможных реализаций схемы базы данных. Вы можете сделать свою, объяснив выбор в комментариях PR)

[//]: # ()
[//]: # (Сперва вам необходимо написать миграции к вашей базе данных.)

[//]: # ()
[//]: # (### Migrations)

[//]: # ()
[//]: # (Создайте директорию [db/migrations]&#40;db/migrations&#41; с вашими миграциями, а)

[//]: # (также [db/migrations/migrate.go]&#40;db/migrations/migrate.go]&#41;)

[//]: # (для их применения)

[//]: # ()
[//]: # (### Author)

[//]: # ()
[//]: # (Создайте таблицу `author`)

[//]: # ()
[//]: # (```sql)

[//]: # (-- +goose Up)

[//]: # (CREATE EXTENSION IF NOT EXISTS "uuid-ossp";)

[//]: # ()
[//]: # (CREATE TABLE author)

[//]: # (&#40;)

[//]: # (    id ...,)

[//]: # (    name ...,)

[//]: # (    created_at ...,)

[//]: # (    updated_at ...)

[//]: # (&#41;;)

[//]: # ()
[//]: # (-- +goose StatementBegin)

[//]: # (CREATE OR REPLACE FUNCTION update_author_timestamp&#40;&#41; RETURNS TRIGGER AS)

[//]: # ($$)

[//]: # (BEGIN)

[//]: # (    NEW.updated_at = now&#40;&#41;;)

[//]: # (    RETURN NEW;)

[//]: # (END;)

[//]: # ($$ LANGUAGE plpgsql;)

[//]: # (-- +goose StatementEnd)

[//]: # ()
[//]: # ()
[//]: # (CREATE OR REPLACE TRIGGER trigger_update_author_timestamp)

[//]: # (    BEFORE UPDATE)

[//]: # (    ON ...)

[//]: # (    FOR EACH ROW)

[//]: # (EXECUTE FUNCTION update_author_timestamp&#40;&#41;;)

[//]: # ()
[//]: # ()
[//]: # (-- +goose Down)

[//]: # (DROP TABLE ...;)

[//]: # (```)

[//]: # ()
[//]: # (Отдельной миграцией создайте индекс на имя автора)

[//]: # ()
[//]: # (```sql)

[//]: # (-- +goose Up)

[//]: # (CREATE INDEX ...;)

[//]: # ()
[//]: # (-- +goose Down)

[//]: # (DROP INDEX ...;)

[//]: # (```)

[//]: # ()
[//]: # (### Book)

[//]: # ()
[//]: # (Создайте таблицу `book`)

[//]: # ()
[//]: # (```sql)

[//]: # (-- +goose Up)

[//]: # (CREATE TABLE book)

[//]: # (&#40;)

[//]: # (    id ...,)

[//]: # (    name ...,)

[//]: # (    created_at ...,)

[//]: # (    updated_at ...)

[//]: # (&#41;;)

[//]: # ()
[//]: # (-- +goose StatementBegin)

[//]: # (CREATE OR REPLACE FUNCTION update_book_timestamp&#40;&#41; RETURNS TRIGGER AS)

[//]: # ($$)

[//]: # (BEGIN)

[//]: # (    NEW.updated_at = now&#40;&#41;;)

[//]: # (    RETURN NEW;)

[//]: # (END;)

[//]: # ($$ LANGUAGE plpgsql;)

[//]: # (-- +goose StatementEnd)

[//]: # ()
[//]: # (CREATE OR REPLACE TRIGGER trigger_update_book_timestamp)

[//]: # (    BEFORE UPDATE)

[//]: # (    ON ...)

[//]: # (    FOR EACH ROW)

[//]: # (EXECUTE FUNCTION update_book_timestamp&#40;&#41;;)

[//]: # ()
[//]: # (-- +goose Down)

[//]: # (DROP TABLE ...;)

[//]: # (```)

[//]: # ()
[//]: # (Отдельной миграцией создайте индекс на имя книги)

[//]: # ()
[//]: # (```sql)

[//]: # (-- +goose Up)

[//]: # (CREATE INDEX  ...;)

[//]: # ()
[//]: # (-- +goose Down)

[//]: # (DROP INDEX ...;)

[//]: # (```)

[//]: # ()
[//]: # (### Book to authors)

[//]: # ()
[//]: # (Создайте таблицу `author_book`)

[//]: # ()
[//]: # (```sql)

[//]: # (-- +goose Up)

[//]: # (CREATE TABLE author_book)

[//]: # (&#40;)

[//]: # (    author_id ...,)

[//]: # (    book_id ...,)

[//]: # (    PRIMARY KEY &#40;.. .&#41;)

[//]: # (&#41;;)

[//]: # ()
[//]: # (-- +goose Down)

[//]: # (DROP TABLE author_book;)

[//]: # (```)

[//]: # ()
[//]: # (- Добавьте `foreign key` для author_id и book_id.)

[//]: # (- Поддержите каскадное удаление `ON DELETE CASCADE`, в случае удаления автора или книги в этой таблице не должны)

[//]: # (  остаться неконсистентные записи)

[//]: # (- Добавьте композитный `PRIMARY KEY`, состоящий из `author_id` и `book_id`)

[//]: # ()
[//]: # (Композитный `PRIMARY KEY` по умолчанию добавляет индекс на свои части, однако)

[//]: # (его [эффективность для каждого атрибута разная]&#40;https://www.postgresql.org/docs/current/indexes-multicolumn.html&#41;.)

[//]: # (Отдельной миграцией добавьте индекс для `book_id`)

[//]: # ()
[//]: # (## Часть 2)

[//]: # ()
[//]: # (В файле [db/migrations/migrate.go]&#40;./db/migrations/migrate.go&#41; напишите код, который будет накатывать миграции.)

[//]: # (Используйте библиотеки ниже, а также `//go:embed migrations/*.sql` для загрузки)

[//]: # (миграций - [пример]&#40;https://github.com/pressly/goose&#41;)

[//]: # ()
[//]: # (```go)

[//]: # ("github.com/jackc/pgx/v5/pgxpool")

[//]: # ("github.com/jackc/pgx/v5/stdlib")

[//]: # ("github.com/pressly/goose/v3")

[//]: # ("github.com/project/library/config")

[//]: # (```)

[//]: # ()
[//]: # (Попробуйте поднять базу данных и проверить, что ваши миграции корректно накатываются)

[//]: # ()
[//]: # (```)

[//]: # (docker volumes)

[//]: # (docker volume ls // если нужно удалить старый volume)

[//]: # (docker volume rm ... // если нужно удалить старый volume)

[//]: # (docker-compose up -d)

[//]: # ()
[//]: # (docker ps -a // посмотреть контейнеры)

[//]: # (docker stop / docker rm - для остановки и удаления контейнера)

[//]: # (```)

[//]: # ()
[//]: # (```)

[//]: # (2025/03/06 15:03:14 OK   001_create_author_table.sql &#40;5.89ms&#41;)

[//]: # (2025/03/06 15:03:14 OK   002_create_author_name_index.sql &#40;8.83ms&#41;)

[//]: # (2025/03/06 15:03:14 OK   003_create_book_table.sql &#40;9.78ms&#41;)

[//]: # (2025/03/06 15:03:14 OK   004_create_book_name_index.sql &#40;2.51ms&#41;)

[//]: # (2025/03/06 15:03:14 OK   005_create_author_book_table.sql &#40;3.28ms&#41;)

[//]: # (2025/03/06 15:03:14 OK   006_create_author_book_book_id_index.sql &#40;2.99ms&#41;)

[//]: # (2025/03/06 15:03:14 goose: successfully migrated database to version: 6)

[//]: # (```)

[//]: # ()
[//]: # (## Часть 3)

[//]: # ()
[//]: # (Поддержите в вашем конфиге параметры для подключения к базе данных)

[//]: # ()
[//]: # (```go)

[//]: # (type &#40;)

[//]: # (    Config struct {)

[//]: # (        GRPC)

[//]: # (        PG)

[//]: # (    })

[//]: # (    )
[//]: # (    GRPC struct {)

[//]: # (        Port        string `env:"GRPC_PORT"`)

[//]: # (        GatewayPort string `env:"GRPC_GATEWAY_PORT"`)

[//]: # (    })

[//]: # (    )
[//]: # (    PG struct {)

[//]: # (        URL      string)

[//]: # (        Host     string `env:"POSTGRES_HOST"`)

[//]: # (        Port     string `env:"POSTGRES_PORT"`)

[//]: # (        DB       string `env:"POSTGRES_DB"`)

[//]: # (        User     string `env:"POSTGRES_USER"`)

[//]: # (        Password string `env:"POSTGRES_PASSWORD"`)

[//]: # (        MaxConn  string `env:"POSTGRES_MAX_CONN"`)

[//]: # (    })

[//]: # (&#41;)

[//]: # (```)

[//]: # ()
[//]: # (Пример URL:)

[//]: # ()
[//]: # (```)

[//]: # (postgres://user:password@host:port/dbname?sslmode=disable&pool_max_conns=10)

[//]: # (```)

[//]: # ()
[//]: # (## Часть 4)

[//]: # ()
[//]: # (Добавьте новую реализацию репозитория вашего сервиса, используя поднятую базу данных.)

[//]: # (Не забывайте про консистентность и атомарность операций. Пример:)

[//]: # ()
[//]: # (```go)

[//]: # (func &#40;r *PostgresRepository&#41; CreateBook&#40;ctx context.Context, book entity.Book&#41; &#40;entity.Book, error&#41; {)

[//]: # (tx, err := r.db.Begin&#40;ctx&#41;)

[//]: # (if err != nil {)

[//]: # (return entity.Book{}, err)

[//]: # (})

[//]: # (defer tx.Rollback&#40;ctx&#41;)

[//]: # ()
[//]: # (	const queryBook = `INSERT INTO book &#40;name&#41; VALUES &#40;$1&#41; RETURNING id, created_at, updated_at`)

[//]: # (	err = tx.QueryRow&#40;ctx, queryBook, book.Name&#41;.Scan&#40;&book.ID, &book.CreatedAt, &book.UpdatedAt&#41;)

[//]: # (	if err != nil {)

[//]: # (		return entity.Book{}, err)

[//]: # (	})

[//]: # ()
[//]: # (	const queryAuthorBooks = `INSERT INTO author_book &#40;author_id, book_id&#41; VALUES &#40;$1, $2&#41;`)

[//]: # (	for _, authorID := range book.AuthorIDs {)

[//]: # (		_, err := tx.Exec&#40;ctx, queryAuthorBooks, authorID, book.ID&#41;)

[//]: # (		if err != nil {)

[//]: # (			return entity.Book{}, err)

[//]: # (		})

[//]: # (	})

[//]: # ()
[//]: # (	if err := tx.Commit&#40;ctx&#41;; err != nil {)

[//]: # (		return entity.Book{}, err)

[//]: # (	})

[//]: # ()
[//]: # (	return book, nil)

[//]: # (})

[//]: # (```)

[//]: # ()
[//]: # (* Старайтесь обойтись одним запросом там, где это возможно)

[//]: # (* ID автора и книги должны генерироваться __на уровне базы__ через `DEFAULT uuid_generate_v4&#40;&#41;`)

[//]: # ()
[//]: # (## Часть 5)

[//]: # ()
[//]: # (Добавьте в API для Book поля `created_at` и `updated_at`)

[//]: # ()
[//]: # (```protobuf)

[//]: # (import "google/protobuf/timestamp.proto";)

[//]: # ()
[//]: # (message Book {)

[//]: # (  ...)

[//]: # (  google.protobuf.Timestamp created_at = ...;)

[//]: # (  google.protobuf.Timestamp updated_at = ...;)

[//]: # (})

[//]: # (```)

[//]: # ()
[//]: # (# HW 2 &#40;outbox&#41;)

[//]: # (## Часть 6)

[//]: # (С этой части начинается ДЗ `outbox`. Ветка с решением должна иметь название `outbox`. Важно, чтобы в PR не было diff'a старого ДЗ.)

[//]: # (Вы можете добиться этого, сделав rebase на `main` после проверки предыдущего ДЗ)

[//]: # ()
[//]: # (Реализуйте паттерн `outbox`, который обсуждался на лекции)

[//]: # ()
[//]: # (Создайте таблицу `outbox`)

[//]: # (```sql)

[//]: # (CREATE TYPE outbox_status as ENUM &#40;'CREATED', 'IN_PROGRESS', 'SUCCESS'&#41;;)

[//]: # ()
[//]: # (CREATE TABLE outbox)

[//]: # (&#40;)

[//]: # (    idempotency_key TEXT PRIMARY KEY,)

[//]: # (    data            JSONB                   NOT NULL,)

[//]: # (    status          outbox_status           NOT NULL,)

[//]: # (    kind            INT                     NOT NULL,)

[//]: # (    created_at      TIMESTAMP DEFAULT now&#40;&#41; NOT NULL,)

[//]: # (    updated_at      TIMESTAMP DEFAULT now&#40;&#41; NOT NULL)

[//]: # (&#41;;)

[//]: # (```)

[//]: # ()
[//]: # (Поддержите транзакции на уровне доменной логики)

[//]: # ()
[//]: # (```go)

[//]: # (type Transactor interface {)

[//]: # (	WithTx&#40;ctx context.Context, function func&#40;ctx context.Context&#41; error&#41; error)

[//]: # (})

[//]: # ()
[//]: # (func extractTx&#40;ctx context.Context&#41; &#40;pgx.Tx, error&#41; {})

[//]: # ()
[//]: # (func injectTx&#40;ctx context.Context, pool *pgxpool.Pool&#41; &#40;context.Context, error, pgx.Tx&#41; {})

[//]: # (```)

[//]: # ()
[//]: # (Например:)

[//]: # (```go)

[//]: # (func &#40;l *libraryImpl&#41; RegisterBook&#40;ctx context.Context, name string, authorIDs []string&#41; &#40;*library.AddBookResponse, error&#41; {)

[//]: # (    var book entity.Book)

[//]: # (	err := l.transactor.WithTx&#40;ctx, func&#40;ctx context.Context&#41; error {)

[//]: # (		book, txErr = l.booksRepository.CreateBook&#40;ctx, entity.Book{)

[//]: # (			Name:      name,)

[//]: # (			AuthorIDs: authorIDs,)

[//]: # (		}&#41;)

[//]: # (		)
[//]: # (		...)

[//]: # (		l.outboxRepository.SendMessage&#40;ctx, idempotencyKey, repository.OutboxKindBook, serialized&#41;)

[//]: # (	}&#41;)

[//]: # (	)
[//]: # (	...)

[//]: # (})

[//]: # (```)

[//]: # ()
[//]: # ()
[//]: # (Поддержите конфиг для Outbox)

[//]: # ()
[//]: # (```go)

[//]: # (type Outbox struct {)

[//]: # (    Enabled         bool          `env:"OUTBOX_ENABLED"`)

[//]: # (    Workers         int           `env:"OUTBOX_WORKERS"`)

[//]: # (    BatchSize       int           `env:"OUTBOX_BATCH_SIZE"`)

[//]: # (    WaitTimeMS      time.Duration `env:"OUTBOX_WAIT_TIME_MS"`)

[//]: # (    InProgressTTLMS time.Duration `env:"OUTBOX_IN_PROGRESS_TTL_MS"`)

[//]: # (    AuthorSendURL   string        `env:"OUTBOX_AUTHOR_SEND_URL"`)

[//]: # (    BookSendURL     string        `env:"OUTBOX_BOOK_SEND_URL"`)

[//]: # (})

[//]: # (```)

[//]: # ()
[//]: # (При создании книги или автора вам необходимо асинхронно отправить `POST` запрос c `AuthorID` или `BookID` на `OUTBOX_AUTHOR_SEND_URL`)

[//]: # (или `OUTBOX_BOOK_SEND_URL`, соответственно.)

[//]: # ()
[//]: # ()
[//]: # (## Унификация технологий)

[//]: # ()
[//]: # (Для удобства выполнения и проверки дз вводится ряд правил, унифицирующих используемые технологии)

[//]: # ()
[//]: # (* Структура проекта [go-clean-template]&#40;https://github.com/evrone/go-clean-template&#41; и)

[//]: # (  этот [шаблон]&#40;https://github.com/itmo-org/lectures/tree/main/sem2/lecture1&#41;)

[//]: # (* Для генерации кода авторские [Makefile]&#40;./Makefile&#41; и [easyp.yaml]&#40;./easyp.yaml&#41;)

[//]: # (* Для логирования [zap]&#40;https://github.com/uber-go/zap&#41;)

[//]: # (* Для валидации [protoc-gen-validate]&#40;https://github.com/bufbuild/protoc-gen-validate&#41;)

[//]: # (* Для поддержики REST-to-gRPC API [gRPC gateway]&#40;https://grpc-ecosystem.github.io/grpc-gateway/&#41;)

[//]: # (* Для миграций [goose]&#40;https://github.com/pressly/goose&#41;)

[//]: # (* [pgx]&#40;https://github.com/jackc/pgx&#41; как драйвер для postgres)

[//]: # ()
[//]: # (## Тестирование в CI)

[//]: # ()
[//]: # (* Код тестов можно посмотреть в файле [integration_test.go]&#40;./integration-test/integration_test.go&#41;)

[//]: # ()
[//]: # (* Важно, чтобы ваш сервис умел корректно обрабатывать SIGINT и SIGTERM, иначе тесты могут работать некорректно)

[//]: # (* В [Makefile]&#40;Makefile&#41; реализованы метки **build** и **generate**, без них CI не будет работать)

[//]: # ()
[//]: # (## Переменные окружения)

[//]: # ()
[//]: # (В рамках вашего сервиса вы должны реализовать конфиг, который будет работать с переменными окружения)

[//]: # ()
[//]: # (## Тесты)

[//]: # ()
[//]: # (Необходимо сгенерировать моки и написать свои тесты, степень покрытия будет проверяться в CI)

[//]: # ()
[//]: # (## Документация)

[//]: # ()
[//]: # (Вам необходимо своими словами написать [README.md]&#40;./docs/README.md&#41; в ./docs к своему сервису library)

[//]: # ()
[//]: # (## Рекомендации)

[//]: # ()
[//]: # (* [Пример реализации]&#40;https://github.com/itmo-org/lectures/tree/main/sem2&#41;)

[//]: # (* Не забывайте про логирование)

[//]: # (* Не забывайте про консистентность в базе данных)

[//]: # (* Используйте [тесты]&#40;./integration-test&#41; чтобы осознать недосказанности)

[//]: # (* Не нужно добавлять старую in-memory реализацию репозитория)

[//]: # ()
[//]: # (## Письменные комментарии)

[//]: # ()
[//]: # (Поскольку количество попыток сдачи ограничено, вы можете написать дополнительные комментарии в PR. Если ваше)

[//]: # ()
[//]: # (обоснование будет достаточно разумным, это может быть учтено при выставлении баллов. Например,)

[//]: # ()
[//]: # (* описать, почему вы написали именно такие интерфейсы)

[//]: # ()
[//]: # (* описать, почему вы сделали именно такую валидацию)

[//]: # ()
[//]: # (* описать, почему вы сделали именно такую схему в базе данных)

[//]: # ()
[//]: # (## Сдача)

[//]: # ()
[//]: # (* Открыть pull request из ветки задания в ветку `main` **вашего репозитория**.)

[//]: # ()
[//]: # (* В описании PR заполнить количество часов, которые вы потратили на это задание.)

[//]: # ()
[//]: # (* Отправить заявку на ревью в соответствующей форме.)

[//]: # ()
[//]: # (* Время дедлайна фиксируется отправкой формы.)

[//]: # ()
[//]: # (* Изменять файлы в ветке main без PR запрещено.)

[//]: # ()
[//]: # (* Изменять файл [CI workflow]&#40;./.github/workflows/library.yaml&#41; запрещено.)

[//]: # ()
[//]: # (## Makefile)

[//]: # ()
[//]: # (Для удобств локальной разработки сделан [`Makefile`]&#40;Makefile&#41;. Имеются следующие команды:)

[//]: # ()
[//]: # (Запустить полный цикл &#40;линтер, тесты&#41;:)

[//]: # ()
[//]: # (```bash )

[//]: # ()
[//]: # (make all)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (Запустить только тесты:)

[//]: # ()
[//]: # (```bash)

[//]: # ()
[//]: # (make test)

[//]: # ()
[//]: # (``` )

[//]: # ()
[//]: # (Запустить линтер:)

[//]: # ()
[//]: # (```bash)

[//]: # ()
[//]: # (make lint)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (Подтянуть новые тесты:)

[//]: # ()
[//]: # (```bash)

[//]: # ()
[//]: # (make update)

[//]: # ()
[//]: # (```)

[//]: # ()
[//]: # (При разработке на Windows рекомендуется использовать [WSL]&#40;https://learn.microsoft.com/en-us/windows/wsl/install&#41;, чтобы)

[//]: # ()
[//]: # (была возможность пользоваться вспомогательными скриптами.)
