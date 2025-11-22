```mermaid
sequenceDiagram
    participant Client
    participant UserService
    Client->>UserService: ListUserServiceRequest
    loop Server Streaming
        UserService-->>Client: Stream ListUserServiceResponse
    end
```
