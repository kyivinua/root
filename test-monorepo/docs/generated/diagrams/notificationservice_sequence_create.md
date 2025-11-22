```mermaid
sequenceDiagram
    participant Client
    participant NotificationService
    Client->>+NotificationService: CreateNotificationServiceRequest
    NotificationService-->>-Client: CreateNotificationServiceResponse
```
