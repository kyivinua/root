```mermaid
graph TD
    AnalyticsService[AnalyticsService]
    AnalyticsService_m0[Create]
    AnalyticsService --> AnalyticsService_m0
    AnalyticsService_m0_in[CreateAnalyticsServiceRequest]
    AnalyticsService_m0_out[CreateAnalyticsServiceResponse]
    AnalyticsService_m0_in -.-> AnalyticsService_m0
    AnalyticsService_m0 -.-> AnalyticsService_m0_out
    AnalyticsService_m1[Get]
    AnalyticsService --> AnalyticsService_m1
    AnalyticsService_m1_in[GetAnalyticsServiceRequest]
    AnalyticsService_m1_out[GetAnalyticsServiceResponse]
    AnalyticsService_m1_in -.-> AnalyticsService_m1
    AnalyticsService_m1 -.-> AnalyticsService_m1_out
    AnalyticsService_m2[List (Server Stream)]
    AnalyticsService --> AnalyticsService_m2
    AnalyticsService_m2_in[ListAnalyticsServiceRequest]
    AnalyticsService_m2_out[ListAnalyticsServiceResponse]
    AnalyticsService_m2_in -.-> AnalyticsService_m2
    AnalyticsService_m2 -.-> AnalyticsService_m2_out
```
