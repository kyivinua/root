```mermaid
sequenceDiagram
    participant Client
    participant UserService
    Client->>+UserService: GetUserServiceRequest
    UserService-->>-Client: GetUserServiceResponse
```
