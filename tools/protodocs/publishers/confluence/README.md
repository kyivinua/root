# Confluence Publisher

Публикует сгенерированную документацию Protocol Buffers в Atlassian Confluence.

## Возможности

- ✅ Автоматическая публикация Markdown документации в Confluence
- ✅ Поддержка создания отдельных страниц для каждого сервиса
- ✅ Поддержка консолидированной страницы для всех сервисов
- ✅ Конвертация Markdown в Confluence Storage Format (HTML)
- ✅ Поддержка Mermaid диаграмм
- ✅ Поддержка PlantUML диаграмм (встроенная поддержка)
- ✅ Автоматическое создание Table of Contents
- ✅ Обновление существующих страниц с версионированием
- ✅ Фильтрация по visibility (PUBLIC, PARTNER, INTERNAL)
- ✅ Поддержка attachments (загрузка и скачивание файлов)
- ✅ Batch operations для ускорения публикации
- ✅ Labels и metadata для страниц
- ✅ Page restrictions (управление доступом)
- ✅ Diff checking для избежания лишних обновлений
- ✅ Retry logic с exponential backoff
- ✅ Content caching для снижения API calls
- ✅ Prometheus metrics для мониторинга
- ✅ Circuit breaker для fault tolerance
- ✅ Space export (полный экспорт пространства в ZIP)

## Конфигурация

### В pipeline.yaml

```yaml
publishers:
  confluence:
    enabled: true
    base_url: "https://company.atlassian.net/wiki"
    username: "${CONFLUENCE_USERNAME}"
    api_token: "${CONFLUENCE_API_TOKEN}"
    space_key: "APIDOCS"
    parent_page_id: "123456789"

    # Структура страниц
    create_page_per_service: true
    page_title_prefix: "API Documentation -"
    include_toc: true
    include_diagrams: true
    include_code_examples: true

    # Поведение при обновлении
    update_existing: true
    version_label: "Auto-generated"

    # Фильтрация контента
    visibility_filter:
      - "PUBLIC"
      - "PARTNER"
```

### Переменные окружения

Рекомендуется использовать переменные окружения для чувствительных данных:

```bash
export CONFLUENCE_USERNAME="your.email@company.com"
export CONFLUENCE_API_TOKEN="your-api-token"
```

### Получение API Token

1. Перейдите в https://id.atlassian.com/manage-profile/security/api-tokens
2. Создайте новый API token
3. Сохраните token в переменную окружения `CONFLUENCE_API_TOKEN`

## Использование

### В составе pipeline

Publisher автоматически запускается на Stage 8 pipeline, если `enabled: true`.

### Standalone использование

```go
package main

import (
    "github.com/kyivinua/docgen-tool/tools/protodocs/publishers/confluence"
)

func main() {
    cfg := &confluence.PublisherConfig{
        BaseURL:              "https://company.atlassian.net/wiki",
        Username:             "your.email@company.com",
        APIToken:             "your-api-token",
        SpaceKey:             "APIDOCS",
        ParentPageID:         "123456789",
        CreatePagePerService: true,
        PageTitlePrefix:      "API -",
        IncludeTOC:           true,
        IncludeDiagrams:      true,
        UpdateExisting:       true,
    }

    publisher := confluence.NewPublisher(cfg)

    // Публикация из директории с Markdown
    result, err := publisher.PublishFromMarkdownFiles("./api-docs/proto-docs")
    if err != nil {
        panic(err)
    }

    fmt.Printf("Created: %d, Updated: %d\n", result.PagesCreated, result.PagesUpdated)
}
```

## Конвертация Markdown → Confluence

Publisher автоматически конвертирует:

| Markdown | Confluence |
|----------|------------|
| `# Header` | `<h1>Header</h1>` |
| ` ```code``` ` | Code macro |
| ` ```mermaid``` ` | Code macro с пометкой |
| `**bold**` | `<strong>bold</strong>` |
| `*italic*` | `<em>italic</em>` |
| `[link](url)` | `<a href="url">link</a>` |
| Tables | `<table>` |
| Lists | `<ul>/<ol>` |
| Blockquotes | Info panel |

## Troubleshooting

### 401 Unauthorized
- Проверьте правильность username и API token
- Убедитесь, что используете API token, а не пароль

### 403 Forbidden
- Проверьте права доступа к Space
- Убедитесь, что у пользователя есть права на создание страниц

### 404 Not Found
- Проверьте правильность `base_url`
- Проверьте существование `space_key`
- Проверьте существование `parent_page_id`

### Diagram не отображаются
- Mermaid диаграммы требуют плагина в Confluence
- Рассмотрите использование [Mermaid for Confluence](https://marketplace.atlassian.com/apps/1224722/mermaid-diagrams-for-confluence)

## Архитектура

```
publishers/confluence/
├── client.go               # Confluence REST API client
├── client_test.go          # API client tests
├── client_extensions.go    # Extended API operations (labels, batch, diff, restrictions)
├── formatter.go            # Markdown → Storage Format converter
├── formatter_test.go       # Formatter tests
├── publisher.go            # Main publisher logic
├── publisher_test.go       # Publisher tests
├── retry.go                # Retry logic with exponential backoff
├── cache.go                # LRU cache with TTL
├── cache_test.go           # Cache tests
├── metrics.go              # Prometheus metrics
├── circuit_breaker.go      # Circuit breaker pattern
├── circuit_breaker_test.go # Circuit breaker tests
├── space_export.go         # Space export to ZIP
└── space_export_test.go    # Space export tests
```

### Core Components

**client.go**
- REST API взаимодействие с Confluence
- CRUD операции для страниц
- Поиск страниц по title
- Attachment upload/download
- Rate limiting (10 req/s)

**formatter.go**
- Конвертация Markdown → Confluence Storage Format
- Обработка кода, таблиц, списков
- Поддержка Mermaid диаграмм
- Поддержка PlantUML диаграмм (нативная)

**publisher.go**
- Оркестрация процесса публикации
- Логика создания/обновления страниц
- Обработка visibility filters

### Extended Features

**client_extensions.go**
- Labels API (добавление, удаление, получение)
- Batch operations (параллельное создание/обновление страниц)
- Diff checking (hash-based для избежания лишних обновлений)
- Page restrictions (read/update permissions для users/groups)

**retry.go**
- Exponential backoff с jitter
- Настраиваемые retry attempts
- Retry только для определенных типов ошибок
- Context cancellation support

**cache.go**
- LRU cache с TTL
- Thread-safe операции
- Background cleanup для expired entries
- PageCache wrapper с typed методами
- Интеграция с метриками

**metrics.go**
- Prometheus metrics
- API request tracking (total, duration, errors)
- Cache performance (hits, misses, evictions)
- Circuit breaker state
- Publisher operations

**circuit_breaker.go**
- Три состояния: Closed, HalfOpen, Open
- Configurable failure thresholds
- Automatic recovery testing
- Failure ratio tracking
- Metrics integration

**space_export.go**
- Экспорт всего Confluence space в ZIP
- Поддержка attachments
- Paginated page retrieval
- Context cancellation
- Export statistics

## Best Practices

1. **Используйте parent_page_id** - группируйте документацию под родительской страницей
2. **Enable update_existing** - для CI/CD автоматизации
3. **Используйте version_label** - отслеживайте версии обновлений
4. **Фильтруйте по visibility** - не публикуйте INTERNAL API в публичное пространство
5. **Храните credentials в env vars** - не коммитьте API tokens в git

## Ограничения

- Mermaid диаграммы сохраняются как код-блоки (требуется плагин для рендеринга)
- PlantUML диаграммы используют встроенный Confluence macro (работает out-of-the-box)
- Rate limiting Confluence API (10 requests/second) - автоматически обрабатывается
- Cache по умолчанию отключен (используйте NewClientWithCache для включения)

## Production Features

### Caching
Используйте кеширование для снижения нагрузки на API:

```go
cacheConfig := &confluence.CacheConfig{
    MaxSize:         1000,             // максимум 1000 страниц в кеше
    TTL:             10 * time.Minute, // время жизни записи
    CleanupInterval: 1 * time.Minute,  // частота очистки
}

client, err := confluence.NewClientWithCache(baseURL, username, apiToken, cacheConfig)
// Автоматически использует cache для GetPage/FindPageByTitle
```

### Metrics
Интегрируйте Prometheus metrics для мониторинга:

```go
metrics := confluence.NewMetrics("confluence_publisher")

// Используйте metrics в клиенте
cacheConfig := &confluence.CacheConfig{
    Metrics: metrics,
}

// Metrics автоматически записываются:
// - confluence_api_requests_total
// - confluence_api_request_duration_seconds
// - confluence_cache_hits_total
// - confluence_cache_misses_total
// - confluence_circuit_breaker_state
```

### Circuit Breaker
Защитите систему от cascading failures:

```go
cbConfig := confluence.DefaultCircuitBreakerConfig("confluence")
cbConfig.MaxFailures = 5
cbConfig.Timeout = 30 * time.Second

cb := confluence.NewCircuitBreaker(cbConfig)

// Используйте circuit breaker для критичных операций
err := cb.Execute(ctx, func(ctx context.Context) error {
    _, err := client.CreatePage(page)
    return err
})
```

### Batch Operations
Ускорьте публикацию с параллельными операциями:

```go
pages := []*confluence.Page{page1, page2, page3}

results, err := client.BatchCreatePages(ctx, pages, 5) // 5 concurrent requests
for pageID, result := range results {
    if result.Error != nil {
        log.Printf("Failed to create page %s: %v", pageID, result.Error)
    }
}
```

### Page Restrictions
Управляйте доступом к страницам:

```go
// Ограничить просмотр для определенных пользователей
err := client.AddPageRestrictions(ctx, pageID,
    confluence.RestrictionOperationRead,
    []string{"user1-id", "user2-id"}, // users
    []string{"developers"},            // groups
)

// Сделать страницу публичной
err := client.MakePagePublic(ctx, pageID)
```

### Space Export
Экспортируйте весь space для бэкапа или миграции:

```go
config := &confluence.SpaceExportConfig{
    SpaceKey:           "APIDOCS",
    OutputPath:         "backup.zip",
    IncludeAttachments: true,
    MaxConcurrency:     5,
    Timeout:            30 * time.Minute,
}

result, err := client.ExportSpace(ctx, config)
fmt.Printf("Exported %d pages, %d attachments (%d bytes) in %v\n",
    result.PagesExported,
    result.AttachmentsExported,
    result.TotalSize,
    result.Duration,
)
```

### Diff Checking
Избегайте лишних обновлений с проверкой изменений:

```go
page, _ := client.FindPageByTitle(spaceKey, title)

hasChanged, err := client.HasContentChanged(page.ID, newContent)
if !hasChanged {
    log.Println("Content unchanged, skipping update")
    return
}

// Обновляем только если есть изменения
updated, err := client.UpdatePage(page.ID, updatedPage)
```

## Advanced Usage

### PlantUML Diagrams
PlantUML диаграммы автоматически конвертируются в Confluence macro:

```markdown
\`\`\`plantuml
@startuml
Alice -> Bob: Hello
Bob -> Alice: Hi!
@enduml
\`\`\`
```

Также поддерживаются другие типы диаграмм:
- `@startmindmap` / `@endmindmap` - mind maps
- `@startgantt` / `@endgantt` - Gantt charts
- `@startsalt` / `@endsalt` - wireframes
- и другие PlantUML типы

### Labels и Metadata
Добавляйте labels для организации документации:

```go
// Добавить labels к странице
err := client.AddLabels(ctx, pageID, []string{"api", "v1", "public"})

// Получить labels
labels, err := client.GetLabels(ctx, pageID)

// Удалить label
err := client.RemoveLabel(ctx, pageID, "draft")
```

## Test Coverage

Проект имеет comprehensive test coverage:

- ✅ API client tests (30+ tests)
- ✅ Formatter tests (20+ tests)
- ✅ Publisher tests (15+ tests)
- ✅ Cache tests (11 tests)
- ✅ Circuit breaker tests (11 tests)
- ✅ Space export tests (8 tests)
- ✅ Retry logic tests

Запустить все тесты:
```bash
go test ./publishers/confluence/... -v
```

## Future Improvements

### Low Priority
- [ ] Add webhook support for page updates
- [ ] Implement page templates
- [ ] Add search API integration
- [ ] Support for Confluence analytics
- [ ] Bulk delete operations
