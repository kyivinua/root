```mermaid
sequenceDiagram
    participant Client
    participant NotificationService
    Client->>+NotificationService: GetNotificationServiceRequest
    NotificationService-->>-Client: GetNotificationServiceResponse
```
