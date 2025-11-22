```mermaid
sequenceDiagram
    participant Client
    participant PaymentService
    Client->>+PaymentService: CreatePaymentServiceRequest
    PaymentService-->>-Client: CreatePaymentServiceResponse
```
