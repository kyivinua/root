# ProtoDocs & ProtoContext System

Комплексное решение для автоматической генерации документации из Protocol Buffer схем и runtime работы с дескрипторами.

## Обзор

Система состоит из двух основных компонентов:

1. **ProtoContext** (`tools/protoctx`) - Go библиотека для runtime работы с Protobuf дескрипторами
2. **ProtoDocsPipeline** (`tools/protodocs`) - Пайплайн автоматической генерации документации

## Архитектура

```
┌─────────────────────────────────────────────────────────────┐
│                    .proto файлы (источник)                   │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│              ProtoDocsPipeline (генерация)                   │
│  ┌───────────┐  ┌──────────┐  ┌────────────┐              │
│  │ Discovery │→ │   Lint   │→ │  Breaking  │              │
│  └───────────┘  └──────────┘  └────────────┘              │
│                       │                                      │
│                       ▼                                      │
│  ┌──────────────────────────────────┐                      │
│  │  Descriptor Build (image.bin)     │                      │
│  └──────────────────────────────────┘                      │
│                       │                                      │
│                       ▼                                      │
│  ┌──────────────────────────────────┐                      │
│  │  Doc Model (ApiDocModel)          │                      │
│  └──────────────────────────────────┘                      │
└─────────────────────────────────────────────────────────────┘
                            │
                            ▼
┌─────────────────────────────────────────────────────────────┐
│                  Runtime Services                            │
│  ┌──────────────────────────────────────┐                  │
│  │  ProtoContext (runtime работа)        │                  │
│  │  - JSON ↔ wire конвертация           │                  │
│  │  - Рефлексия типов                   │                  │
│  │  - Извлечение комментариев            │                  │
│  │  - Проверка совместимости             │                  │
│  └──────────────────────────────────────┘                  │
└─────────────────────────────────────────────────────────────┘
```

## ProtoContext

### Возможности

- **Загрузка дескрипторов** из FileDescriptorSet (image.bin)
- **Рефлексия типов**: поиск и перечисление сервисов, сообщений, enum
- **JSON ↔ wire конвертация** для любых типов
- **Динамические сообщения** без сгенерированного Go кода
- **Извлечение комментариев** из SourceCodeInfo
- **Проверка совместимости** между двумя версиями схем

### Использование

```go
import "github.com/kyivinua/root/tools/protoctx"

// Загрузка контекста
ctx, err := protoctx.LoadFromDescriptorSetFile("api-docs/descriptors/image.bin")
if err != nil {
    log.Fatal(err)
}

// Поиск дескриптора сервиса
sd, err := ctx.FindServiceDescriptor("company.user.v1.UserService")

// Конвертация wire → JSON
jsonData, err := ctx.MarshalWireToJSON("company.user.v1.UserProfile", wireBytes)

// Конвертация JSON → wire
wireBytes, err := ctx.UnmarshalJSONToWire("company.user.v1.UserProfile", jsonData)

// Получение комментариев
comments, _ := ctx.CommentsIndex().GetByFQN("company.user.v1.UserService")
fmt.Println(comments.Summary())

// Проверка совместимости
report, err := newCtx.CompareSchemas(oldCtx, protoctx.DefaultCompatibilityRules())
if report.HasBreaking() {
    // Обработка breaking changes
}
```

## ProtoDocsPipeline

### Стадии пайплайна

1. **Discovery** - определение изменённых .proto файлов
2. **Lint** - проверка структуры и стиля (buf lint)
3. **Breaking Check** - проверка обратной совместимости
4. **Descriptor Build** - сборка image.bin (FileDescriptorSet)
5. **Doc Model Build** - построение ApiDocModel из дескрипторов
6. **Docs Generation** - генерация Markdown/HTML
7. **OpenAPI Generation** - генерация OpenAPI спецификаций
8. **Site Assembly** - сборка статического сайта

### Использование через CLI

```bash
# Запуск полного пайплайна
go run ./cmd/proto-docs all --config configs/proto-docs.config.yaml

# Или через Makefile
make proto-docs

# Только линтинг
make proto-lint

# Только сборка дескрипторов
make proto-build

# Проверка breaking changes
make proto-breaking
```

## Runtime Service

Пример HTTP сервиса, использующего ProtoContext для runtime операций.

### Endpoints

- `GET /debug/readyz` - проверка готовности
- `GET /debug/schema/service?name=<FQN>` - получение схемы сервиса
- `POST /proto/decode?type=<FQN>` - декодирование wire → JSON
- `POST /proto/encode?type=<FQN>` - кодирование JSON → wire

### Запуск

```bash
# Сборка дескрипторов
make proto-build

# Запуск runtime сервиса
make run-runtime

# В другом терминале:
# Получение схемы сервиса
curl http://localhost:8080/debug/schema/service?name=company.user.v1.UserService

# Декодирование wire в JSON
curl -X POST -H "Content-Type: application/octet-stream" \
  --data-binary @user.wire \
  http://localhost:8080/proto/decode?type=company.user.v1.UserProfile
```

## Структура проекта

```
/
├── proto/                    # Исходники .proto
│   ├── user/v1/
│   │   ├── user_service.proto
│   │   └── user_types.proto
│   └── ...
├── tools/
│   ├── protoctx/             # ProtoContext библиотека
│   │   ├── context.go
│   │   ├── loader.go
│   │   ├── file_index.go
│   │   ├── comments.go
│   │   ├── json_codec.go
│   │   ├── reflection.go
│   │   └── compat.go
│   └── protodocs/            # Pipeline
│       └── pipeline/
│           ├── config.go
│           ├── discovery.go
│           ├── model.go
│           └── pipeline.go
├── cmd/
│   ├── proto-docs/           # CLI для пайплайна
│   │   └── main.go
│   └── runtime/              # Runtime сервис
│       ├── main.go
│       └── state.go
├── configs/
│   ├── buf.yaml
│   └── proto-docs.config.yaml
└── api-docs/                 # Генерируемые артефакты (в .gitignore)
    ├── descriptors/
    │   └── image.bin
    ├── model/
    │   └── api-doc-model.json
    ├── proto-docs/
    └── openapi/
```

## Конфигурация

### configs/proto-docs.config.yaml

```yaml
proto_root: "./proto"
use_buf: true

lint:
  enable_buf_lint: true
  enable_comments_check: true

breaking:
  enable: true
  target: ".git#branch=main"

descriptors:
  output_path: "api-docs/descriptors/image.bin"

docs:
  output_format: "markdown"
  output_dir: "./api-docs/proto-docs"
  visibility_filter: ["PUBLIC", "PARTNER"]

openapi:
  enabled: true
  plugin: "openapiv3"
  output_dir: "./api-docs/openapi"
  visibility_filter: ["PUBLIC"]
```

### configs/buf.yaml

```yaml
version: v1
name: buf.build/company/monorepo

build:
  roots:
    - ../proto

lint:
  use:
    - DEFAULT
```

## Требования

- Go 1.22+
- Protocol Buffers
- Buf CLI (опционально, для lint/breaking/build)

## Установка зависимостей

```bash
# Go модули
go mod download

# Buf CLI (для lint/breaking/build)
# macOS
brew install bufbuild/buf/buf

# Linux
# Скачать с https://github.com/bufbuild/buf/releases
```

## Сборка

```bash
# Сборка всех компонентов
make build-proto-docs
make build-runtime

# Или просто
make build
```

## Тестирование

```bash
# Все тесты
make test

# Только ProtoContext
make test-protoctx

# Только Pipeline
make test-pipeline
```

## Примеры .proto файлов

### proto/user/v1/user_service.proto

```protobuf
syntax = "proto3";

package company.user.v1;

option go_package = "github.com/kyivinua/root/proto/user/v1;userv1";

import "user/v1/user_types.proto";

// Сервис управления пользователями.
service UserService {
  // Создает нового пользователя.
  rpc CreateUser(CreateUserRequest) returns (CreateUserResponse) {}

  // Получает профиль пользователя.
  rpc GetUser(GetUserRequest) returns (GetUserResponse) {}
}

message CreateUserRequest {
  // Email-адрес пользователя.
  string email = 1;

  // Имя пользователя для отображения.
  string display_name = 2;
}

message CreateUserResponse {
  // Профиль созданного пользователя.
  UserProfile user = 1;
}
```

## Модель документации (ApiDocModel)

Система строит иерархическую модель документации:

```
ApiDocModel
  ├── Modules[] (пакеты)
  │   ├── Services[]
  │   │   └── Methods[]
  │   ├── Messages[]
  │   │   └── Fields[]
  │   └── Enums[]
  │       └── Values[]
  └── Statistics
```

Модель сохраняется в JSON: `api-docs/model/api-doc-model.json`

## Workflow в CI/CD

```yaml
# .github/workflows/proto-docs.yml
name: Proto Docs

on:
  pull_request:
    paths:
      - "proto/**"
  push:
    branches: [main]

jobs:
  proto-docs:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: "1.22"

      - name: Run proto CI
        run: make proto-ci

      - name: Upload docs
        uses: actions/upload-artifact@v4
        with:
          name: proto-docs
          path: api-docs/
```

## Лицензия

См. LICENSE

## Авторы

- kyivinua

## Ссылки

- [Protocol Buffers Documentation](https://protobuf.dev/)
- [Buf Documentation](https://buf.build/docs)
- [Google API Design Guide](https://cloud.google.com/apis/design)
