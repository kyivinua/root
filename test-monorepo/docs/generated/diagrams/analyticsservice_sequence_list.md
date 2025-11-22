```mermaid
sequenceDiagram
    participant Client
    participant AnalyticsService
    Client->>AnalyticsService: ListAnalyticsServiceRequest
    loop Server Streaming
        AnalyticsService-->>Client: Stream ListAnalyticsServiceResponse
    end
```
