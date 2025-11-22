```mermaid
sequenceDiagram
    participant Client
    participant NotificationService
    Client->>NotificationService: ListNotificationServiceRequest
    loop Server Streaming
        NotificationService-->>Client: Stream ListNotificationServiceResponse
    end
```
