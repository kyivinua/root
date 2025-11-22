```mermaid
sequenceDiagram
    participant Client
    participant AnalyticsService
    Client->>+AnalyticsService: CreateAnalyticsServiceRequest
    AnalyticsService-->>-Client: CreateAnalyticsServiceResponse
```
