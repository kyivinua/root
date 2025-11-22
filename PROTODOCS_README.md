# ProtoDocs & ProtoContext System

Комплексное решение для автоматической генерации документации из Protocol Buffer схем и runtime работы с дескрипторами.

## Обзор

Система состоит из трех основных компонентов:

1. **ProtoContext** (`tools/protoctx`) - Go библиотека для runtime работы с Protobuf дескрипторами
2. **ProtoDocsPipeline** (`tools/protodocs`) - Пайплайн автоматической генерации документации
3. **Enrichment Module** (`tools/protodocs/enricher`) - LLM-based обогащение документации с RAG, safety guards и audit trail

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
6. **Enrichment** (опционально) - LLM-based обогащение документации
7. **Docs Generation** - генерация Markdown/HTML
8. **OpenAPI Generation** - генерация OpenAPI спецификаций
9. **Site Assembly** - сборка статического сайта

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

## Enrichment Module

Enterprise-grade LLM-based documentation enrichment with RAG, hallucination detection, and full audit trail.

### Возможности

- **Multi-Provider LLM Support**: Anthropic Claude, OpenAI, Ollama через gollm
- **RAG Integration**: Weaviate vector store + langchaingo для контекстной генерации
- **Semantic Entropy**: Обнаружение галлюцинаций через von Neumann entropy
- **LLM-as-Judge**: Верификация достоверности через второй LLM вызов
- **Safety Guards**: PII/PCI-DSS проверки, regex-based фильтры
- **Advanced Prompting**: XML tags + Chain-of-Thought для Claude
- **Adaptive RAG**: Умное использование RAG на основе сложности задачи
- **Tenant Isolation**: Policy engine с поддержкой мультиарендности
- **Full Audit Trail**: Полный трейс каждого enrichment с токенами, стоимостью, временем
- **Metrics & Monitoring**: Prometheus метрики для production deployment
- **Caching**: Ristretto cache для снижения затрат
- **Enrichment Manifest**: Отчет о результатах для site assembly

### Архитектура

```
┌────────────────┐
│  ApiDocModel   │
└────────┬───────┘
         │
         ▼
┌────────────────────────────────────────┐
│         Enricher Orchestrator          │
│  ┌──────────┐  ┌──────────┐           │
│  │ Policy   │  │  Safety  │           │
│  │ Engine   │  │  Guard   │           │
│  └──────────┘  └──────────┘           │
│         │                              │
│         ▼                              │
│  ┌──────────────────────┐             │
│  │   Smart Strategy     │             │
│  │  (Adaptive RAG)      │             │
│  └──────────────────────┘             │
│         │                              │
│         ▼                              │
│  ┌─────────────┐  ┌─────────────┐    │
│  │ RAG Retriev │  │ LLM Client  │    │
│  │ (Weaviate)  │  │  (gollm)    │    │
│  └─────────────┘  └─────────────┘    │
│         │              │               │
│         └──────┬───────┘               │
│                ▼                        │
│  ┌──────────────────────┐             │
│  │  Template Engine     │             │
│  │  (XML/CoT)           │             │
│  └──────────────────────┘             │
└────────────────────────────────────────┘
         │
         ▼
┌────────────────────────────────────────┐
│   Enriched ApiDocModel + Manifest      │
└────────────────────────────────────────┘
```

### Использование

```bash
# Базовая конфигурация
export LLM_API_KEY="your-anthropic-key"
export WEAVIATE_API_KEY="your-weaviate-key"  # если используется

# Запуск enrichment standalone
make run-enricher

# Или через pipeline (включить в configs/proto-docs.config.yaml)
enrichment:
  enabled: true
  config_path: "configs/enricher.config.yaml"

make proto-docs
```

### Конфигурация

`configs/enricher.config.yaml`:

```yaml
provider: anthropic
model: claude-3-5-sonnet-20241022
temperature: 0.0
max_tokens: 4096

rag:
  enabled: true
  vector_store: weaviate
  use_adaptive: true  # Адаптивное использование RAG

safety:
  enabled: true
  use_semantic_entropy: true  # von Neumann entropy
  use_llm_judge: true         # LLM-as-judge validation
  pii_checks: true
  pci_dss_checks: true

cache:
  enabled: true
  ttl: 24h

metrics:
  enabled: true
  type: prometheus
```

### Enrichment Flow

1. **Policy Check**: Проверка tenant policy (allow/deny/approval required)
2. **Cache Lookup**: Проверка кэша для избежания повторных вызовов
3. **Smart Strategy**: Решение использовать RAG или base LLM
4. **RAG Retrieval** (если нужно): Извлечение top-K релевантных документов
5. **Prompt Rendering**: XML/CoT template с context
6. **LLM Generation**: Вызов LLM с temperature=0.0
7. **Safety Validation**:
   - PII/PCI regex checks
   - Semantic entropy calculation
   - LLM-as-judge faithfulness check
8. **Apply & Cache**: Сохранение enriched docs + кэширование
9. **Audit Trail**: Запись полного trace с метриками

### Safety Guards

#### PII Detection
- Email addresses
- Phone numbers
- SSN patterns
- IP addresses

#### PCI-DSS Detection
- Credit card patterns
- CVV/CVC codes

#### Semantic Entropy
```
entropy = -Σ p(word) * log2(p(word))
normalized_entropy = entropy / log2(unique_words)
```

#### LLM-as-Judge
```
VERDICT: [FAITHFUL|UNFAITHFUL]
CONFIDENCE: [0.0-1.0]
REASONING: [explanation]
```

### Enrichment Manifest

После enrichment генерируется manifest:

```json
{
  "version": "1.0",
  "model_used": "claude-3-5-sonnet-20241022",
  "provider_used": "anthropic",
  "start_time": "2025-01-20T10:00:00Z",
  "duration": "5m30s",
  "statistics": {
    "total_targets": 150,
    "enriched_targets": 148,
    "failed_targets": 2,
    "cache_hits": 45,
    "total_tokens_used": 125000,
    "success_rate": 0.9867,
    "cache_hit_rate": 0.30
  }
}
```

### Makefile Targets

```bash
# Сборка enricher binary
make build-enricher

# Запуск enrichment
make run-enricher

# Тесты enricher
make test-enricher
```

## Notifications & Release Notes

Система поддерживает автоматические уведомления в Slack о статусе pipeline и публикацию release notes.

### Slack Integration

#### Возможности

- **Pipeline Notifications**: Уведомления о старте, завершении и ошибках pipeline
- **Enrichment Results**: Статистика LLM enrichment с метриками
- **Breaking Changes Alerts**: Автоматическое оповещение о breaking changes
- **Release Notes**: Автоматическая генерация и публикация release notes
- **Rich Formatting**: Цветные сообщения с полями и метриками

#### Конфигурация

```yaml
# configs/proto-docs.config.yaml
notifications:
  enabled: true
  slack:
    enabled: true
    # webhook_url: ""  # или через SLACK_WEBHOOK_URL env var
    # bot_token: ""    # альтернатива webhook через SLACK_BOT_TOKEN
    # channel: "#proto-docs"
    username: "ProtoDocs Bot"
    icon_emoji: ":book:"
    notify_on_start: true
    notify_on_complete: true
    notify_on_failure: true
    notify_on_breaking: true
    notify_on_enrichment: true
    notify_release_notes: true
    release_notes_version: "v1.0.0"
```

#### Использование

```bash
# Настройка через Webhook (рекомендуется)
export SLACK_WEBHOOK_URL="https://hooks.slack.com/services/YOUR/WEBHOOK/URL"

# Или через Bot Token (для более сложных сценариев)
export SLACK_BOT_TOKEN="xoxb-your-bot-token"
export SLACK_CHANNEL="#proto-docs"

# Запуск pipeline с уведомлениями
make proto-docs
```

### Release Notes Generator

Автоматическая генерация release notes из git commits с поддержкой Conventional Commits.

#### Поддерживаемые типы коммитов

- `feat:` → ✨ New Features
- `fix:` → 🐛 Bug Fixes
- `perf:`/`refactor:`/`improvement:` → 🔧 Improvements
- `feat!:` или `BREAKING CHANGE:` → ⚠️ Breaking Changes
- `deprecate:` → 🗑️ Deprecations

#### Пример Release Notes в Slack

```
📝 Release Notes - v1.0.0

✨ New Features:
• Add Enrichment Module with LLM support
• Implement RAG integration with Weaviate

🔧 Improvements:
• Optimize ProtoContext loading performance
• Enhance safety guards with entropy detection

Statistics:
• Services: 15
• Messages: 120
• LLM Enriched Targets: 148
• Success Rate: 98.7%
```

#### Автоматическая публикация

```yaml
# Включить публикацию release notes
notifications:
  slack:
    notify_release_notes: true
    release_notes_version: "v1.1.0"  # Текущая версия
    release_notes_from_ref: "v1.0.0"  # Сравнить с этим ref
```

### Типы уведомлений

| Тип | Описание | Цвет |
|-----|----------|------|
| Pipeline Start | Старт документации build | Зеленый |
| Pipeline Complete | Успешное завершение | Зеленый |
| Pipeline Failed | Ошибка pipeline | Красный |
| Enrichment Complete | Результаты LLM enrichment | Желтый/Зеленый |
| Breaking Changes | Обнаружены breaking changes | Красный |
| Release Notes | Публикация release notes | Синий |

## Mermaid Architecture Diagrams

Система автоматически генерирует Mermaid диаграммы для визуализации архитектуры, компонентов и структуры данных.

### Возможности

- **Pipeline Architecture**: Flowchart всех стадий pipeline с decision points
- **Enricher Orchestration**: Детальная схема LLM enrichment с safety guards
- **Component Interaction**: Граф зависимостей всех модулей и внешних сервисов
- **Proto → Doc Transformation**: Визуализация преобразования .proto в документацию
- **Deployment Architecture**: Схема runtime deployment с конфигурациями
- **Data Model ER**: Entity-Relationship диаграмма структуры ApiDocModel
- **Service Map**: Граф связей между сервисами и сообщениями
- **Message Hierarchy**: Дерево зависимостей message types

### Конфигурация

```yaml
# configs/proto-docs.config.yaml
diagrams:
  enabled: true
  output_dir: "./api-docs/diagrams"

  # Включить/выключить отдельные типы диаграмм
  enable_pipeline: true           # Pipeline architecture flowchart
  enable_enricher: true           # Enricher orchestration flowchart
  enable_component: true          # Component interaction graph
  enable_transform: true          # Proto → Doc transformation flowchart
  enable_deploy: true             # Deployment architecture graph
  enable_data_model: true         # Data model ER diagram
  enable_service_map: true        # Service relationship map
  enable_message_hierarchy: false # Message hierarchy (может быть очень большой)

  # Настройки генерации
  generate_index: true            # Генерировать README.md index
  theme: "default"                # Mermaid тема: default, forest, dark, neutral
  max_services_per_diagram: 20    # Лимит сервисов в service map
  max_messages_per_diagram: 30    # Лимит сообщений в message hierarchy
```

### Использование

```bash
# Генерация диаграмм вместе с pipeline
make proto-docs

# Диаграммы будут созданы в api-docs/diagrams/
ls -la api-docs/diagrams/
# pipeline-architecture.md
# enricher-orchestration.md
# component-interaction.md
# proto-transform.md
# deployment-architecture.md
# data-model-structure.md
# service-map.md
# README.md  # Индекс всех диаграмм
```

### Типы диаграмм

| Диаграмма | Тип Mermaid | Описание | Динамическая |
|-----------|-------------|----------|--------------|
| Pipeline Architecture | flowchart LR | Полный pipeline со всеми стадиями | Нет |
| Enricher Orchestration | flowchart TD | Процесс LLM enrichment с safety guards | Нет |
| Component Interaction | graph TB | Зависимости между модулями | Нет |
| Proto → Doc Transform | flowchart LR | Преобразование .proto в ApiDocModel | Нет |
| Deployment Architecture | graph LR | Runtime deployment и интеграции | Нет |
| Data Model ER | erDiagram | Структура ApiDocModel | Да |
| Service Map | graph TB | Связи сервисов и сообщений | Да |
| Message Hierarchy | graph TD | Дерево зависимостей messages | Да |

**Статические диаграммы** генерируются один раз на основе архитектуры системы.

**Динамические диаграммы** создаются из ApiDocModel и отражают реальную структуру ваших .proto файлов.

### Интеграция в документацию

```markdown
# В вашем mkdocs.yml или docusaurus.config.js

# Mermaid diagrams автоматически рендерятся браузером:
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>

# Или используйте markdown plugin:
# mkdocs.yml
markdown_extensions:
  - pymdownx.superfences:
      custom_fences:
        - name: mermaid
          class: mermaid
          format: !!python/name:pymdownx.superfences.fence_code_format
```

### Пример диаграммы Pipeline

```mermaid
flowchart LR
    Start([Start Pipeline]) --> Discovery[🔍 Discovery Stage]
    Discovery --> Lint[✓ Lint Stage]
    Lint --> Breaking[⚠️ Breaking Check]
    Breaking --> Build[🔨 Descriptor Build]
    Build --> LoadContext[📥 Load ProtoContext]
    LoadContext --> BuildModel[🏗️ Build Doc Model]
    BuildModel --> EnrichDecision{Enrichment Enabled?}
    EnrichDecision -->|Yes| Enrich[🤖 LLM Enrichment]
    EnrichDecision -->|No| GenDocs[📝 Generate Docs]
    Enrich --> GenDocs
    GenDocs --> End([✅ Complete])
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
