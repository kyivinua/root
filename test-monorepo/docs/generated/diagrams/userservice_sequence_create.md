```mermaid
sequenceDiagram
    participant Client
    participant UserService
    Client->>+UserService: CreateUserServiceRequest
    UserService-->>-Client: CreateUserServiceResponse
```
