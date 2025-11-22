```mermaid
graph TD
    NotificationService[NotificationService]
    NotificationService_m0[Create]
    NotificationService --> NotificationService_m0
    NotificationService_m0_in[CreateNotificationServiceRequest]
    NotificationService_m0_out[CreateNotificationServiceResponse]
    NotificationService_m0_in -.-> NotificationService_m0
    NotificationService_m0 -.-> NotificationService_m0_out
    NotificationService_m1[Get]
    NotificationService --> NotificationService_m1
    NotificationService_m1_in[GetNotificationServiceRequest]
    NotificationService_m1_out[GetNotificationServiceResponse]
    NotificationService_m1_in -.-> NotificationService_m1
    NotificationService_m1 -.-> NotificationService_m1_out
    NotificationService_m2[List (Server Stream)]
    NotificationService --> NotificationService_m2
    NotificationService_m2_in[ListNotificationServiceRequest]
    NotificationService_m2_out[ListNotificationServiceResponse]
    NotificationService_m2_in -.-> NotificationService_m2
    NotificationService_m2 -.-> NotificationService_m2_out
```
