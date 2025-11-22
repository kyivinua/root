```mermaid
graph TD
    PaymentService[PaymentService]
    PaymentService_m0[Create]
    PaymentService --> PaymentService_m0
    PaymentService_m0_in[CreatePaymentServiceRequest]
    PaymentService_m0_out[CreatePaymentServiceResponse]
    PaymentService_m0_in -.-> PaymentService_m0
    PaymentService_m0 -.-> PaymentService_m0_out
    PaymentService_m1[Get]
    PaymentService --> PaymentService_m1
    PaymentService_m1_in[GetPaymentServiceRequest]
    PaymentService_m1_out[GetPaymentServiceResponse]
    PaymentService_m1_in -.-> PaymentService_m1
    PaymentService_m1 -.-> PaymentService_m1_out
    PaymentService_m2[List (Server Stream)]
    PaymentService --> PaymentService_m2
    PaymentService_m2_in[ListPaymentServiceRequest]
    PaymentService_m2_out[ListPaymentServiceResponse]
    PaymentService_m2_in -.-> PaymentService_m2
    PaymentService_m2 -.-> PaymentService_m2_out
```
