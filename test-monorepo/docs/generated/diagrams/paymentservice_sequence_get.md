```mermaid
sequenceDiagram
    participant Client
    participant PaymentService
    Client->>+PaymentService: GetPaymentServiceRequest
    PaymentService-->>-Client: GetPaymentServiceResponse
```
