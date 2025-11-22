# Confluence Publisher

Публикует сгенерированную документацию Protocol Buffers в Atlassian Confluence.

## Возможности

- ✅ Автоматическая публикация Markdown документации в Confluence
- ✅ Поддержка создания отдельных страниц для каждого сервиса
- ✅ Поддержка консолидированной страницы для всех сервисов
- ✅ Конвертация Markdown в Confluence Storage Format (HTML)
- ✅ Поддержка Mermaid диаграмм
- ✅ Автоматическое создание Table of Contents
- ✅ Обновление существующих страниц с версионированием
- ✅ Фильтрация по visibility (PUBLIC, PARTNER, INTERNAL)

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
├── client.go      # Confluence REST API client
├── formatter.go   # Markdown → Storage Format converter
└── publisher.go   # Main publisher logic
```

### client.go
- REST API взаимодействие с Confluence
- CRUD операции для страниц
- Поиск страниц по title

### formatter.go
- Конвертация Markdown → Confluence Storage Format
- Обработка кода, таблиц, списков
- Специальная обработка Mermaid диаграмм

### publisher.go
- Оркестрация процесса публикации
- Логика создания/обновления страниц
- Обработка visibility filters

## Best Practices

1. **Используйте parent_page_id** - группируйте документацию под родительской страницей
2. **Enable update_existing** - для CI/CD автоматизации
3. **Используйте version_label** - отслеживайте версии обновлений
4. **Фильтруйте по visibility** - не публикуйте INTERNAL API в публичное пространство
5. **Храните credentials в env vars** - не коммитьте API tokens в git

## Ограничения

- Mermaid диаграммы сохраняются как код-блоки (требуется плагин для рендеринга)
- Сложные HTML в Markdown может потребовать доработки formatter
- Rate limiting Confluence API (10 requests/second)

## Future Improvements

- [ ] Поддержка attachments (изображения, файлы)
- [ ] Batch operations для ускорения публикации
- [ ] Поддержка labels и metadata
- [ ] Экспорт в Confluence Space format
- [ ] Интеграция с PlantUML для диаграмм
