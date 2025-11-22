```mermaid
sequenceDiagram
    participant Client
    participant AnalyticsService
    Client->>+AnalyticsService: GetAnalyticsServiceRequest
    AnalyticsService-->>-Client: GetAnalyticsServiceResponse
```
