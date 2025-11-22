```mermaid
graph TD
    UserService[UserService]
    UserService_m0[Create]
    UserService --> UserService_m0
    UserService_m0_in[CreateUserServiceRequest]
    UserService_m0_out[CreateUserServiceResponse]
    UserService_m0_in -.-> UserService_m0
    UserService_m0 -.-> UserService_m0_out
    UserService_m1[Get]
    UserService --> UserService_m1
    UserService_m1_in[GetUserServiceRequest]
    UserService_m1_out[GetUserServiceResponse]
    UserService_m1_in -.-> UserService_m1
    UserService_m1 -.-> UserService_m1_out
    UserService_m2[List (Server Stream)]
    UserService --> UserService_m2
    UserService_m2_in[ListUserServiceRequest]
    UserService_m2_out[ListUserServiceResponse]
    UserService_m2_in -.-> UserService_m2
    UserService_m2 -.-> UserService_m2_out
```
