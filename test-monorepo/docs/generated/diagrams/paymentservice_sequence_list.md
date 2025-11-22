```mermaid
sequenceDiagram
    participant Client
    participant PaymentService
    Client->>PaymentService: ListPaymentServiceRequest
    loop Server Streaming
        PaymentService-->>Client: Stream ListPaymentServiceResponse
    end
```
