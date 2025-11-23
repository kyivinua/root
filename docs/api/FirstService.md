# 📚 FirstService API Documentation

| **Attribute** | **Value** |
|---------------|----------|
| **Service Name** | `FirstService` |
| **Package** | `first.v1` |
| **Version** | v1 |
| **Proto File** | `first/first.proto` |
| **Generated** | 2025-11-23T01:12:47Z |

FirstService manages projects and tasks with full lifecycle operations.

---

## 📑 Table of Contents

- [Overview](#overview)
- [Authentication & Authorization](#authentication)
- [Rate Limits & Quotas](#rate-limits)
- [Service Level Agreement (SLA)](#sla)
- [API Versioning & Lifecycle](#versioning)
- [Architecture](#architecture)
- [gRPC Service Interactions](#service-interaction)
- [Message Type Diagrams](#class-diagram)
- [Methods](#methods)
  - [CreateProject](#createproject)
  - [GetProject](#getproject)
  - [UpdateProject](#updateproject)
  - [DeleteProject](#deleteproject)
  - [ListProjects](#listprojects)
  - [AddProjectMember](#addprojectmember)
  - [RemoveProjectMember](#removeprojectmember)
  - [CreateTask](#createtask)
  - [GetTask](#gettask)
  - [UpdateTask](#updatetask)
  - [DeleteTask](#deletetask)
  - [ListTasks](#listtasks)
  - [AssignTask](#assigntask)
  - [CompleteTask](#completetask)
  - [StreamProjectUpdates](#streamprojectupdates)
  - [BatchCreateTasks](#batchcreatetasks)
  - [SyncProjectData](#syncprojectdata)
- [Messages](#messages)
  - [GetProjectResponse](#getprojectresponse)
  - [AddProjectMemberResponse](#addprojectmemberresponse)
  - [CreateTaskRequest](#createtaskrequest)
  - [ListTasksRequest](#listtasksrequest)
  - [ProjectSyncMessage](#projectsyncmessage)
  - [DeleteProjectResponse](#deleteprojectresponse)
  - [UpdateTaskResponse](#updatetaskresponse)
  - [ProjectUpdate](#projectupdate)
  - [AssignTaskResponse](#assigntaskresponse)
  - [CompleteTaskRequest](#completetaskrequest)
  - [StreamProjectUpdatesRequest](#streamprojectupdatesrequest)
  - [RemoveProjectMemberRequest](#removeprojectmemberrequest)
  - [GetTaskResponse](#gettaskresponse)
  - [ListTasksResponse](#listtasksresponse)
  - [BatchCreateTasksResponse](#batchcreatetasksresponse)
  - [CreateProjectResponse](#createprojectresponse)
  - [GetProjectRequest](#getprojectrequest)
  - [UpdateProjectRequest](#updateprojectrequest)
  - [UpdateProjectResponse](#updateprojectresponse)
  - [GetTaskRequest](#gettaskrequest)
  - [DeleteTaskRequest](#deletetaskrequest)
  - [DeleteProjectRequest](#deleteprojectrequest)
  - [ListProjectsResponse](#listprojectsresponse)
  - [AddProjectMemberRequest](#addprojectmemberrequest)
  - [UpdateTaskRequest](#updatetaskrequest)
  - [CompleteTaskResponse](#completetaskresponse)
  - [CreateProjectRequest](#createprojectrequest)
  - [RemoveProjectMemberResponse](#removeprojectmemberresponse)
  - [ListProjectsRequest](#listprojectsrequest)
  - [CreateTaskResponse](#createtaskresponse)
  - [DeleteTaskResponse](#deletetaskresponse)
  - [AssignTaskRequest](#assigntaskrequest)
- [Error Codes](#error-codes)
- [Examples](#examples)

---

## 📖 Overview

<a name="overview"></a>

### Service Statistics

| Metric | Count |
|--------|-------|
| **RPC Methods** | 17 |
| **Message Types** | 32 |
| **Enumerations** | 0 |
| **Streaming RPCs** | 3 |

### Quick Start

This service provides the following capabilities:

- [`CreateProject`](#createproject): CreateProject creates a new project.
- [`GetProject`](#getproject): GetProject retrieves a project by ID.
- [`UpdateProject`](#updateproject): UpdateProject updates an existing project.
- [`DeleteProject`](#deleteproject): DeleteProject soft-deletes a project.
- [`ListProjects`](#listprojects): ListProjects lists projects with pagination and filtering.
- ... and 12 more methods

---

## 🔐 Authentication & Authorization

<a name="authentication"></a>

### Supported Authentication Methods

This service supports the following authentication methods:

1. **API Keys** - For service-to-service communication
2. **OAuth 2.0** - For user-delegated access with bearer tokens
3. **JWT Tokens** - For stateless authentication

### Authentication Examples

#### Using API Keys

```bash
# Command-line (grpcurl)
grpcurl -H 'x-api-key: YOUR_API_KEY' \
  api.example.com:443 \
  first.v1.FirstService/CreateProject
```

#### Using OAuth 2.0 Bearer Token (Go)

```go
import (
    "context"
    "google.golang.org/grpc"
    "google.golang.org/grpc/metadata"
)

func callWithAuth(client pb.ServiceClient, token string) error {
    // Create metadata with authorization header
    md := metadata.New(map[string]string{
        "authorization": "Bearer " + token,
    })
    ctx := metadata.NewOutgoingContext(context.Background(), md)

    // Make RPC call with authenticated context
    resp, err := client.SomeMethod(ctx, &pb.Request{})
    return err
}
```

#### Using OAuth 2.0 Bearer Token (TypeScript)

```typescript
import * as grpc from '@grpc/grpc-js';

// Create metadata with authorization header
const metadata = new grpc.Metadata();
metadata.add('authorization', 'Bearer ' + accessToken);

// Make RPC call with authenticated metadata
client.someMethod(request, metadata, (error, response) => {
    if (error) {
        console.error('Error:', error);
        return;
    }
    console.log('Response:', response);
});
```

### Authorization Scopes

Different methods require different permission scopes:

| Method | Required Scope | Description |
|--------|---------------|-------------|
| `CreateProject` | `first.write` | Create new resources |
| `GetProject` | `first.read` | Read individual resources |
| `UpdateProject` | `first.write` | Modify existing resources |
| `DeleteProject` | `first.admin` | Administrative access required |
| `ListProjects` | `first.read` | List and query resources |
| `AddProjectMember` | `first.access` | Access resources |
| `RemoveProjectMember` | `first.access` | Access resources |
| `CreateTask` | `first.write` | Create new resources |
| `GetTask` | `first.read` | Read individual resources |
| `UpdateTask` | `first.write` | Modify existing resources |
| `DeleteTask` | `first.admin` | Administrative access required |
| `ListTasks` | `first.read` | List and query resources |
| `AssignTask` | `first.access` | Access resources |
| `CompleteTask` | `first.access` | Access resources |
| `StreamProjectUpdates` | `first.access` | Access resources |
| `BatchCreateTasks` | `first.access` | Access resources |
| `SyncProjectData` | `first.access` | Access resources |

### Security Best Practices

- ✅ Always use TLS 1.3+ in production environments
- ✅ Rotate API keys every 90 days
- ✅ Use short-lived tokens (1 hour maximum)
- ✅ Implement request signing for sensitive operations
- ✅ Store credentials securely (use secret management systems)
- ❌ Never log authentication credentials or tokens
- ❌ Never commit API keys to version control

---

## ⏱️ Rate Limits & Quotas

<a name="rate-limits"></a>

### Standard Rate Limits

The following rate limits apply to all API requests:

| Tier | Requests/Second | Requests/Day | Burst |
|------|----------------|--------------|-------|
| **Free** | 10 | 10000 | 20 |
| **Professional** | 100 | 1000000 | 200 |
| **Enterprise** | 1000 | Unlimited | 2000 |

### Rate Limit Headers

All API responses include rate limit information in the response metadata:

```http
x-ratelimit-limit: 100
x-ratelimit-remaining: 87
x-ratelimit-reset: 1634567890
x-ratelimit-retry-after: 42
```

### Handling Rate Limits

When you exceed rate limits, you'll receive a `RESOURCE_EXHAUSTED` error. Implement exponential backoff to handle rate limiting gracefully:

#### Exponential Backoff Example (TypeScript)

```typescript
async function callWithRetry(
  fn: () => Promise<any>,
  maxRetries = 3
): Promise<any> {
  for (let i = 0; i < maxRetries; i++) {
    try {
      return await fn();
    } catch (error: any) {
      if (error.code === grpc.status.RESOURCE_EXHAUSTED) {
        const delay = Math.min(1000 * Math.pow(2, i), 30000);
        console.log(`Rate limited. Retrying in ${delay}ms...`);
        await new Promise(resolve => setTimeout(resolve, delay));
        continue;
      }
      throw error;
    }
  }
  throw new Error('Max retries exceeded');
}
```

#### Exponential Backoff Example (Go)

```go
import (
    "context"
    "time"
    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
)

func callWithRetry(ctx context.Context, fn func() error, maxRetries int) error {
    for i := 0; i < maxRetries; i++ {
        err := fn()
        if err == nil {
            return nil
        }

        if status.Code(err) == codes.ResourceExhausted {
            delay := time.Duration(1000*math.Pow(2, float64(i))) * time.Millisecond
            if delay > 30*time.Second {
                delay = 30 * time.Second
            }
            log.Printf("Rate limited. Retrying in %v...", delay)
            time.Sleep(delay)
            continue
        }
        return err
    }
    return fmt.Errorf("max retries exceeded")
}
```

---

## 📊 Service Level Agreement (SLA)

<a name="sla"></a>

### Availability Commitments

| Service Tier | Uptime SLA | Monthly Downtime | Latency (p95) |
|-------------|-----------|------------------|---------------|
| **Standard** | 99.9% | 43.8 minutes | < 200ms |
| **Premium** | 99.95% | 21.9 minutes | < 100ms |
| **Enterprise** | 99.99% | 4.38 minutes | < 50ms |

### Performance SLOs

#### Latency Targets (p95)

| Operation Type | Target | Notes |
|---------------|--------|-------|
| **Read Operations** (Get*) | < 100ms | Measured server-side |
| **Write Operations** (Create*, Update*) | < 500ms | Includes validation |
| **List Operations** (List*) | < 200ms | With pagination |
| **Delete Operations** (Delete*) | < 300ms | Soft delete |
| **Streaming** | < 50ms | Time to first message |

### Monitoring & Observability

All services expose Prometheus-compatible metrics:

```prometheus
# Request latency histogram (seconds)
api_request_duration_seconds{service="FirstService",method="CreateProject",quantile="0.95"}

# Total request count
api_requests_total{service="FirstService",method="CreateProject",status="success"}

# Error count by code
api_errors_total{service="FirstService",method="CreateProject",code="INVALID_ARGUMENT"}
```

### Health Check

Health status is available via the standard gRPC health check protocol:

```bash
grpc_health_probe -addr=api.example.com:443 \
  -service=first.v1.FirstService
```

---

## 🔄 API Versioning & Lifecycle

<a name="versioning"></a>

### Versioning Strategy

**Strategy**: Semantic Versioning (package.v{major})

This service follows semantic versioning:

- **Major version** (v1, v2): Breaking changes requiring client updates
- **Minor version** (implicit): Backward-compatible additions
- **Patch version**: Bug fixes (not reflected in package name)

**Current Version**: v1

### Version Support Policy

| Version State | Support Period | Updates | Deprecation Notice |
|--------------|----------------|---------|-------------------|
| **Current** | Indefinite | Features + Fixes | N/A |
| **Previous** | 12 months for previous version | Security fixes only | 6 months advance notice prior |
| **Deprecated** | 6 months | Critical security only | 12 months prior |
| **Sunset** | 0 months | None | Service disabled |

### Breaking vs. Non-Breaking Changes

**Breaking changes** (require major version bump):
- ❌ Removing or renaming services, methods, or fields
- ❌ Changing field types or field numbers
- ❌ Changing method behavior significantly
- ❌ Removing enum values

**Non-breaking changes** (allowed in current version):
- ✅ Adding new services, methods, or fields
- ✅ Adding optional fields
- ✅ Adding enum values
- ✅ Deprecating (but not removing) fields

### Migration Support

When major version updates are released, we provide:

- ✅ Comprehensive migration guides
- ✅ Code examples showing before/after
- ✅ Side-by-side running period (overlap)
- ✅ Automated migration tools (where possible)
- ✅ Dedicated support during migration period

---

## 🏗️ Architecture

<a name="architecture"></a>

```mermaid
%{init: {'theme':'forest'}}%
graph TB
    classDef serviceClass fill:#e1f5ff,stroke:#01579b,stroke-width:3px
    classDef methodClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef messageClass fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px

    FirstService[🔧 FirstService]:::serviceClass

    CreateProject[CreateProject]:::methodClass
    FirstService --> CreateProject
    CreateProject_in[📥 CreateProjectRequest]:::messageClass
    CreateProject_out[📤 CreateProjectResponse]:::messageClass
    CreateProject_in -.->|input| CreateProject
    CreateProject -.->|output| CreateProject_out
    GetProject[GetProject]:::methodClass
    FirstService --> GetProject
    GetProject_in[📥 GetProjectRequest]:::messageClass
    GetProject_out[📤 GetProjectResponse]:::messageClass
    GetProject_in -.->|input| GetProject
    GetProject -.->|output| GetProject_out
    UpdateProject[UpdateProject]:::methodClass
    FirstService --> UpdateProject
    UpdateProject_in[📥 UpdateProjectRequest]:::messageClass
    UpdateProject_out[📤 UpdateProjectResponse]:::messageClass
    UpdateProject_in -.->|input| UpdateProject
    UpdateProject -.->|output| UpdateProject_out
    DeleteProject[DeleteProject]:::methodClass
    FirstService --> DeleteProject
    DeleteProject_in[📥 DeleteProjectRequest]:::messageClass
    DeleteProject_out[📤 DeleteProjectResponse]:::messageClass
    DeleteProject_in -.->|input| DeleteProject
    DeleteProject -.->|output| DeleteProject_out
    ListProjects[ListProjects]:::methodClass
    FirstService --> ListProjects
    ListProjects_in[📥 ListProjectsRequest]:::messageClass
    ListProjects_out[📤 ListProjectsResponse]:::messageClass
    ListProjects_in -.->|input| ListProjects
    ListProjects -.->|output| ListProjects_out
    AddProjectMember[AddProjectMember]:::methodClass
    FirstService --> AddProjectMember
    AddProjectMember_in[📥 AddProjectMemberRequest]:::messageClass
    AddProjectMember_out[📤 AddProjectMemberResponse]:::messageClass
    AddProjectMember_in -.->|input| AddProjectMember
    AddProjectMember -.->|output| AddProjectMember_out
    RemoveProjectMember[RemoveProjectMember]:::methodClass
    FirstService --> RemoveProjectMember
    RemoveProjectMember_in[📥 RemoveProjectMemberRequest]:::messageClass
    RemoveProjectMember_out[📤 RemoveProjectMemberResponse]:::messageClass
    RemoveProjectMember_in -.->|input| RemoveProjectMember
    RemoveProjectMember -.->|output| RemoveProjectMember_out
    CreateTask[CreateTask]:::methodClass
    FirstService --> CreateTask
    CreateTask_in[📥 CreateTaskRequest]:::messageClass
    CreateTask_out[📤 CreateTaskResponse]:::messageClass
    CreateTask_in -.->|input| CreateTask
    CreateTask -.->|output| CreateTask_out
    GetTask[GetTask]:::methodClass
    FirstService --> GetTask
    GetTask_in[📥 GetTaskRequest]:::messageClass
    GetTask_out[📤 GetTaskResponse]:::messageClass
    GetTask_in -.->|input| GetTask
    GetTask -.->|output| GetTask_out
    UpdateTask[UpdateTask]:::methodClass
    FirstService --> UpdateTask
    UpdateTask_in[📥 UpdateTaskRequest]:::messageClass
    UpdateTask_out[📤 UpdateTaskResponse]:::messageClass
    UpdateTask_in -.->|input| UpdateTask
    UpdateTask -.->|output| UpdateTask_out
    DeleteTask[DeleteTask]:::methodClass
    FirstService --> DeleteTask
    DeleteTask_in[📥 DeleteTaskRequest]:::messageClass
    DeleteTask_out[📤 DeleteTaskResponse]:::messageClass
    DeleteTask_in -.->|input| DeleteTask
    DeleteTask -.->|output| DeleteTask_out
    ListTasks[ListTasks]:::methodClass
    FirstService --> ListTasks
    ListTasks_in[📥 ListTasksRequest]:::messageClass
    ListTasks_out[📤 ListTasksResponse]:::messageClass
    ListTasks_in -.->|input| ListTasks
    ListTasks -.->|output| ListTasks_out
    AssignTask[AssignTask]:::methodClass
    FirstService --> AssignTask
    AssignTask_in[📥 AssignTaskRequest]:::messageClass
    AssignTask_out[📤 AssignTaskResponse]:::messageClass
    AssignTask_in -.->|input| AssignTask
    AssignTask -.->|output| AssignTask_out
    CompleteTask[CompleteTask]:::methodClass
    FirstService --> CompleteTask
    CompleteTask_in[📥 CompleteTaskRequest]:::messageClass
    CompleteTask_out[📤 CompleteTaskResponse]:::messageClass
    CompleteTask_in -.->|input| CompleteTask
    CompleteTask -.->|output| CompleteTask_out
    StreamProjectUpdates[↓ StreamProjectUpdates]:::methodClass
    FirstService --> StreamProjectUpdates
    StreamProjectUpdates_in[📥 StreamProjectUpdatesRequest]:::messageClass
    StreamProjectUpdates_out[📤 ProjectUpdate]:::messageClass
    StreamProjectUpdates_in -.->|input| StreamProjectUpdates
    StreamProjectUpdates -.->|output| StreamProjectUpdates_out
    BatchCreateTasks[↑ BatchCreateTasks]:::methodClass
    FirstService --> BatchCreateTasks
    BatchCreateTasks_in[📥 CreateTaskRequest]:::messageClass
    BatchCreateTasks_out[📤 BatchCreateTasksResponse]:::messageClass
    BatchCreateTasks_in -.->|input| BatchCreateTasks
    BatchCreateTasks -.->|output| BatchCreateTasks_out
    SyncProjectData[↔️ SyncProjectData]:::methodClass
    FirstService --> SyncProjectData
    SyncProjectData_in[📥 ProjectSyncMessage]:::messageClass
    SyncProjectData_out[📤 ProjectSyncMessage]:::messageClass
    SyncProjectData_in -.->|input| SyncProjectData
    SyncProjectData -.->|output| SyncProjectData_out
```

---

## 🔄 gRPC Service Interactions

<a name="service-interaction"></a>

This diagram shows the interactions between the service methods and message types.

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant FirstService
    Client->>+FirstService: CreateProject
    Note right of FirstService: CreateProjectRequest
    FirstService->>-Client: CreateProjectResponse
    Client->>+FirstService: GetProject
    Note right of FirstService: GetProjectRequest
    FirstService->>-Client: GetProjectResponse
    Client->>+FirstService: UpdateProject
    Note right of FirstService: UpdateProjectRequest
    FirstService->>-Client: UpdateProjectResponse
    Client->>+FirstService: DeleteProject
    Note right of FirstService: DeleteProjectRequest
    FirstService->>-Client: DeleteProjectResponse
    Client->>+FirstService: ListProjects
    Note right of FirstService: ListProjectsRequest
    FirstService->>-Client: ListProjectsResponse
    Client->>+FirstService: AddProjectMember
    Note right of FirstService: AddProjectMemberRequest
    FirstService->>-Client: AddProjectMemberResponse
    Client->>+FirstService: RemoveProjectMember
    Note right of FirstService: RemoveProjectMemberRequest
    FirstService->>-Client: RemoveProjectMemberResponse
    Client->>+FirstService: CreateTask
    Note right of FirstService: CreateTaskRequest
    FirstService->>-Client: CreateTaskResponse
    Client->>+FirstService: GetTask
    Note right of FirstService: GetTaskRequest
    FirstService->>-Client: GetTaskResponse
    Client->>+FirstService: UpdateTask
    Note right of FirstService: UpdateTaskRequest
    FirstService->>-Client: UpdateTaskResponse
    Client->>+FirstService: DeleteTask
    Note right of FirstService: DeleteTaskRequest
    FirstService->>-Client: DeleteTaskResponse
    Client->>+FirstService: ListTasks
    Note right of FirstService: ListTasksRequest
    FirstService->>-Client: ListTasksResponse
    Client->>+FirstService: AssignTask
    Note right of FirstService: AssignTaskRequest
    FirstService->>-Client: AssignTaskResponse
    Client->>+FirstService: CompleteTask
    Note right of FirstService: CompleteTaskRequest
    FirstService->>-Client: CompleteTaskResponse
    Client->>+FirstService: StreamProjectUpdates
    Note right of FirstService: StreamProjectUpdatesRequest
    FirstService->>-Client: Stream of ProjectUpdate
    Client->>+FirstService: BatchCreateTasks (client stream)
    Note over Client,FirstService: Stream of CreateTaskRequest
    FirstService->>-Client: BatchCreateTasksResponse
    Client->>+FirstService: SyncProjectData (bidirectional stream)
    Note over Client,FirstService: Stream of ProjectSyncMessage
    FirstService->>-Client: Stream of ProjectSyncMessage
```

---

## 📦 Message Type Diagrams

<a name="class-diagram"></a>

UML class diagrams showing the structure of message types.

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class FirstService {
        <<service>>
        +GetProjectResponse()
        +AddProjectMemberResponse()
        +CreateTaskRequest()
        +ListTasksRequest()
        +ProjectSyncMessage()
        +DeleteProjectResponse()
        +UpdateTaskResponse()
        +ProjectUpdate()
        +AssignTaskResponse()
        +CompleteTaskRequest()
        +StreamProjectUpdatesRequest()
        +RemoveProjectMemberRequest()
        +GetTaskResponse()
        +ListTasksResponse()
        +BatchCreateTasksResponse()
        +CreateProjectResponse()
        +GetProjectRequest()
        +UpdateProjectRequest()
        +UpdateProjectResponse()
        +GetTaskRequest()
        +DeleteTaskRequest()
        +DeleteProjectRequest()
        +ListProjectsResponse()
        +AddProjectMemberRequest()
        +UpdateTaskRequest()
        +CompleteTaskResponse()
        +CreateProjectRequest()
        +RemoveProjectMemberResponse()
        +ListProjectsRequest()
        +CreateTaskResponse()
        +DeleteTaskResponse()
        +AssignTaskRequest()
    }

    class GetProjectResponse {
        +Project project
    }

    GetProjectResponse "1" --> "1" Project
    class AddProjectMemberResponse {
        +ProjectMember member
        +Project project
    }

    AddProjectMemberResponse "1" --> "1" ProjectMember
    AddProjectMemberResponse "1" --> "1" Project
    class CreateTaskRequest {
        +string project_id
        +string title
        +string description
        +string assignee_id
        +Priority priority
        +Timestamp due_date
        +double estimated_hours
        +string parent_task_id
        +string dependency_ids[]
        +string labels[]
    }

    CreateTaskRequest "1" --> "1" Priority
    class ListTasksRequest {
        +PaginationRequest pagination
        +string project_id
        +string assignee_id
        +Status status
        +Priority priority
        +ResourceState state
        +string labels[]
        +string search_query
        +Timestamp due_before
        +Timestamp due_after
        +string sort_by
        +bool sort_desc
    }

    ListTasksRequest "1" --> "1" PaginationRequest
    ListTasksRequest "1" --> "1" Status
    ListTasksRequest "1" --> "1" Priority
    ListTasksRequest "1" --> "1" ResourceState
    class ProjectSyncMessage {
        +string message_id
        +string message_type
        +Timestamp timestamp
        +string project_id
        +Project project
        +Task tasks[]
        +string sync_status
        +string error
    }

    ProjectSyncMessage "1" --> "1" Project
    ProjectSyncMessage "1" --> "*" Task
    class DeleteProjectResponse {
        +bool success
        +Timestamp deleted_at
    }

    class UpdateTaskResponse {
        +Task task
    }

    UpdateTaskResponse "1" --> "1" Task
    class ProjectUpdate {
        +Timestamp timestamp
        +string update_type
        +Project project
        +Task task
        +string description
        +string updated_by
    }

    ProjectUpdate "1" --> "1" Project
    ProjectUpdate "1" --> "1" Task
    class AssignTaskResponse {
        +Task task
    }

    AssignTaskResponse "1" --> "1" Task
    class CompleteTaskRequest {
        +string task_id
        +double actual_hours
        +string notes
    }

    class StreamProjectUpdatesRequest {
        +string project_id
        +bool include_tasks
        +bool include_members
    }

    class RemoveProjectMemberRequest {
        +string project_id
        +string user_id
    }

    class GetTaskResponse {
        +Task task
    }

    GetTaskResponse "1" --> "1" Task
    class ListTasksResponse {
        +Task tasks[]
        +PaginationResponse pagination
    }

    ListTasksResponse "1" --> "*" Task
    ListTasksResponse "1" --> "1" PaginationResponse
    class BatchCreateTasksResponse {
        +Task tasks[]
        +int32 created_count
        +int32 failed_count
        +Error errors[]
    }

    BatchCreateTasksResponse "1" --> "*" Task
    BatchCreateTasksResponse "1" --> "*" Error
    class CreateProjectResponse {
        +Project project
    }

    CreateProjectResponse "1" --> "1" Project
    class GetProjectRequest {
        +string project_id
        +bool include_deleted
    }

    class UpdateProjectRequest {
        +string project_id
        +string name
        +string description
        +Timestamp end_date
        +Money budget
        +Status status
        +string tags[]
        +int64 version
    }

    UpdateProjectRequest "1" --> "1" Money
    UpdateProjectRequest "1" --> "1" Status
    class UpdateProjectResponse {
        +Project project
    }

    UpdateProjectResponse "1" --> "1" Project
    class GetTaskRequest {
        +string task_id
        +bool include_deleted
    }

    class DeleteTaskRequest {
        +string task_id
        +bool hard_delete
    }

    class DeleteProjectRequest {
        +string project_id
        +bool hard_delete
    }

    class ListProjectsResponse {
        +Project projects[]
        +PaginationResponse pagination
    }

    ListProjectsResponse "1" --> "*" Project
    ListProjectsResponse "1" --> "1" PaginationResponse
    class AddProjectMemberRequest {
        +string project_id
        +string user_id
        +string role
        +AccessLevel access_level
    }

    AddProjectMemberRequest "1" --> "1" AccessLevel
    class UpdateTaskRequest {
        +string task_id
        +string title
        +string description
        +string assignee_id
        +Priority priority
        +Status status
        +Timestamp due_date
        +double actual_hours
        +int64 version
    }

    UpdateTaskRequest "1" --> "1" Priority
    UpdateTaskRequest "1" --> "1" Status
    class CompleteTaskResponse {
        +Task task
        +Timestamp completed_at
    }

    CompleteTaskResponse "1" --> "1" Task
    class CreateProjectRequest {
        +string name
        +string description
        +string owner_id
        +Timestamp start_date
        +Timestamp end_date
        +Money budget
        +ProjectMember members[]
        +string tags[]
    }

    CreateProjectRequest "1" --> "1" Money
    CreateProjectRequest "1" --> "*" ProjectMember
    class RemoveProjectMemberResponse {
        +bool success
        +Project project
    }

    RemoveProjectMemberResponse "1" --> "1" Project
    class ListProjectsRequest {
        +PaginationRequest pagination
        +string owner_id
        +ResourceState state
        +Status status
        +string tags[]
        +string search_query
        +string sort_by
        +bool sort_desc
    }

    ListProjectsRequest "1" --> "1" PaginationRequest
    ListProjectsRequest "1" --> "1" ResourceState
    ListProjectsRequest "1" --> "1" Status
    class CreateTaskResponse {
        +Task task
    }

    CreateTaskResponse "1" --> "1" Task
    class DeleteTaskResponse {
        +bool success
        +Timestamp deleted_at
    }

    class AssignTaskRequest {
        +string task_id
        +string assignee_id
    }

```

---

## ⚙️ Methods

<a name="methods"></a>

This service defines **17 RPC methods**:

### CreateProject

<a name="createproject"></a>

CreateProject creates a new project.

#### Method Signature

```protobuf
// Unary RPC
rpc CreateProject(CreateProjectRequest) returns (CreateProjectResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateProject` |
| **Input Type** | [`CreateProjectRequest`](#createprojectrequest) |
| **Output Type** | [`CreateProjectResponse`](#createprojectresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CreateProject
    Note right of Service: CreateProjectRequest
    Service-->>-Client: Response
    Note left of Client: CreateProjectResponse
```

---

### GetProject

<a name="getproject"></a>

GetProject retrieves a project by ID.

#### Method Signature

```protobuf
// Unary RPC
rpc GetProject(GetProjectRequest) returns (GetProjectResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetProject` |
| **Input Type** | [`GetProjectRequest`](#getprojectrequest) |
| **Output Type** | [`GetProjectResponse`](#getprojectresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetProject
    Note right of Service: GetProjectRequest
    Service-->>-Client: Response
    Note left of Client: GetProjectResponse
```

---

### UpdateProject

<a name="updateproject"></a>

UpdateProject updates an existing project.

#### Method Signature

```protobuf
// Unary RPC
rpc UpdateProject(UpdateProjectRequest) returns (UpdateProjectResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateProject` |
| **Input Type** | [`UpdateProjectRequest`](#updateprojectrequest) |
| **Output Type** | [`UpdateProjectResponse`](#updateprojectresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: UpdateProject
    Note right of Service: UpdateProjectRequest
    Service-->>-Client: Response
    Note left of Client: UpdateProjectResponse
```

---

### DeleteProject

<a name="deleteproject"></a>

DeleteProject soft-deletes a project.

#### Method Signature

```protobuf
// Unary RPC
rpc DeleteProject(DeleteProjectRequest) returns (DeleteProjectResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteProject` |
| **Input Type** | [`DeleteProjectRequest`](#deleteprojectrequest) |
| **Output Type** | [`DeleteProjectResponse`](#deleteprojectresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: DeleteProject
    Note right of Service: DeleteProjectRequest
    Service-->>-Client: Response
    Note left of Client: DeleteProjectResponse
```

---

### ListProjects

<a name="listprojects"></a>

ListProjects lists projects with pagination and filtering.

#### Method Signature

```protobuf
// Unary RPC
rpc ListProjects(ListProjectsRequest) returns (ListProjectsResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListProjects` |
| **Input Type** | [`ListProjectsRequest`](#listprojectsrequest) |
| **Output Type** | [`ListProjectsResponse`](#listprojectsresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: ListProjects
    Note right of Service: ListProjectsRequest
    Service-->>-Client: Response
    Note left of Client: ListProjectsResponse
```

---

### AddProjectMember

<a name="addprojectmember"></a>

AddProjectMember adds a member to a project.

#### Method Signature

```protobuf
// Unary RPC
rpc AddProjectMember(AddProjectMemberRequest) returns (AddProjectMemberResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AddProjectMember` |
| **Input Type** | [`AddProjectMemberRequest`](#addprojectmemberrequest) |
| **Output Type** | [`AddProjectMemberResponse`](#addprojectmemberresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: AddProjectMember
    Note right of Service: AddProjectMemberRequest
    Service-->>-Client: Response
    Note left of Client: AddProjectMemberResponse
```

---

### RemoveProjectMember

<a name="removeprojectmember"></a>

RemoveProjectMember removes a member from a project.

#### Method Signature

```protobuf
// Unary RPC
rpc RemoveProjectMember(RemoveProjectMemberRequest) returns (RemoveProjectMemberResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.RemoveProjectMember` |
| **Input Type** | [`RemoveProjectMemberRequest`](#removeprojectmemberrequest) |
| **Output Type** | [`RemoveProjectMemberResponse`](#removeprojectmemberresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: RemoveProjectMember
    Note right of Service: RemoveProjectMemberRequest
    Service-->>-Client: Response
    Note left of Client: RemoveProjectMemberResponse
```

---

### CreateTask

<a name="createtask"></a>

CreateTask creates a new task in a project.

#### Method Signature

```protobuf
// Unary RPC
rpc CreateTask(CreateTaskRequest) returns (CreateTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateTask` |
| **Input Type** | [`CreateTaskRequest`](#createtaskrequest) |
| **Output Type** | [`CreateTaskResponse`](#createtaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CreateTask
    Note right of Service: CreateTaskRequest
    Service-->>-Client: Response
    Note left of Client: CreateTaskResponse
```

---

### GetTask

<a name="gettask"></a>

GetTask retrieves a task by ID.

#### Method Signature

```protobuf
// Unary RPC
rpc GetTask(GetTaskRequest) returns (GetTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetTask` |
| **Input Type** | [`GetTaskRequest`](#gettaskrequest) |
| **Output Type** | [`GetTaskResponse`](#gettaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetTask
    Note right of Service: GetTaskRequest
    Service-->>-Client: Response
    Note left of Client: GetTaskResponse
```

---

### UpdateTask

<a name="updatetask"></a>

UpdateTask updates an existing task.

#### Method Signature

```protobuf
// Unary RPC
rpc UpdateTask(UpdateTaskRequest) returns (UpdateTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateTask` |
| **Input Type** | [`UpdateTaskRequest`](#updatetaskrequest) |
| **Output Type** | [`UpdateTaskResponse`](#updatetaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: UpdateTask
    Note right of Service: UpdateTaskRequest
    Service-->>-Client: Response
    Note left of Client: UpdateTaskResponse
```

---

### DeleteTask

<a name="deletetask"></a>

DeleteTask soft-deletes a task.

#### Method Signature

```protobuf
// Unary RPC
rpc DeleteTask(DeleteTaskRequest) returns (DeleteTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteTask` |
| **Input Type** | [`DeleteTaskRequest`](#deletetaskrequest) |
| **Output Type** | [`DeleteTaskResponse`](#deletetaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: DeleteTask
    Note right of Service: DeleteTaskRequest
    Service-->>-Client: Response
    Note left of Client: DeleteTaskResponse
```

---

### ListTasks

<a name="listtasks"></a>

ListTasks lists tasks with pagination and filtering.

#### Method Signature

```protobuf
// Unary RPC
rpc ListTasks(ListTasksRequest) returns (ListTasksResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListTasks` |
| **Input Type** | [`ListTasksRequest`](#listtasksrequest) |
| **Output Type** | [`ListTasksResponse`](#listtasksresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: ListTasks
    Note right of Service: ListTasksRequest
    Service-->>-Client: Response
    Note left of Client: ListTasksResponse
```

---

### AssignTask

<a name="assigntask"></a>

AssignTask assigns a task to a user.

#### Method Signature

```protobuf
// Unary RPC
rpc AssignTask(AssignTaskRequest) returns (AssignTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AssignTask` |
| **Input Type** | [`AssignTaskRequest`](#assigntaskrequest) |
| **Output Type** | [`AssignTaskResponse`](#assigntaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: AssignTask
    Note right of Service: AssignTaskRequest
    Service-->>-Client: Response
    Note left of Client: AssignTaskResponse
```

---

### CompleteTask

<a name="completetask"></a>

CompleteTask marks a task as completed.

#### Method Signature

```protobuf
// Unary RPC
rpc CompleteTask(CompleteTaskRequest) returns (CompleteTaskResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CompleteTask` |
| **Input Type** | [`CompleteTaskRequest`](#completetaskrequest) |
| **Output Type** | [`CompleteTaskResponse`](#completetaskresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CompleteTask
    Note right of Service: CompleteTaskRequest
    Service-->>-Client: Response
    Note left of Client: CompleteTaskResponse
```

---

### StreamProjectUpdates

<a name="streamprojectupdates"></a>

StreamProjectUpdates streams real-time updates for a project. Uses server-side streaming.

#### Method Signature

```protobuf
// Server streaming RPC
rpc StreamProjectUpdates(StreamProjectUpdatesRequest) returns (stream ProjectUpdate);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.StreamProjectUpdates` |
| **Input Type** | [`StreamProjectUpdatesRequest`](#streamprojectupdatesrequest) |
| **Output Type** | [`ProjectUpdate`](#projectupdate) |
| **Streaming Type** | Server Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Server Streaming
    Client->>+Service: StreamProjectUpdates
    Client->>Service: StreamProjectUpdatesRequest
    loop Stream Messages
        Service-->>Client: ProjectUpdate
    end
    Service-->>-Client: End Stream
```

---

### BatchCreateTasks

<a name="batchcreatetasks"></a>

BatchCreateTasks creates multiple tasks. Uses client-side streaming.

#### Method Signature

```protobuf
// Client streaming RPC
rpc BatchCreateTasks(stream CreateTaskRequest) returns (BatchCreateTasksResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.BatchCreateTasks` |
| **Input Type** | [`CreateTaskRequest`](#createtaskrequest) |
| **Output Type** | [`BatchCreateTasksResponse`](#batchcreatetasksresponse) |
| **Streaming Type** | Client Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Client Streaming
    Client->>+Service: BatchCreateTasks (stream)
    loop Stream Messages
        Client->>Service: CreateTaskRequest
    end
    Service-->>-Client: BatchCreateTasksResponse
```

---

### SyncProjectData

<a name="syncprojectdata"></a>

SyncProjectData synchronizes project data bidirectionally. Uses bidirectional streaming.

#### Method Signature

```protobuf
// Bidirectional streaming RPC
rpc SyncProjectData(stream ProjectSyncMessage) returns (stream ProjectSyncMessage);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.SyncProjectData` |
| **Input Type** | [`ProjectSyncMessage`](#projectsyncmessage) |
| **Output Type** | [`ProjectSyncMessage`](#projectsyncmessage) |
| **Streaming Type** | Bidirectional Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Bidirectional Streaming
    Client->>+Service: SyncProjectData (stream)
    loop Stream Messages
        Client->>Service: ProjectSyncMessage
        Service-->>Client: ProjectSyncMessage
    end
    Service-->>-Client: End Stream
```

---

## 📦 Messages

<a name="messages"></a>

This service defines **32 message types**:

### GetProjectResponse

<a name="getprojectresponse"></a>

GetProjectResponse returns the requested project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetProjectResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project` | [`Project`](#project) | optional | Retrieved project. |

#### Proto Definition

```protobuf
message GetProjectResponse {
  // Retrieved project.
  optional Project project = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetProjectResponse {
        +Project project
    }
    GetProjectResponse --> Project
```

---

### AddProjectMemberResponse

<a name="addprojectmemberresponse"></a>

AddProjectMemberResponse returns the added member.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AddProjectMemberResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `member` | [`ProjectMember`](#projectmember) | optional | Added member. |
| 2 | `project` | [`Project`](#project) | optional | Updated project. |

#### Proto Definition

```protobuf
message AddProjectMemberResponse {
  // Added member.
  optional ProjectMember member = 1;
  // Updated project.
  optional Project project = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AddProjectMemberResponse {
        +ProjectMember member
        +Project project
    }
    AddProjectMemberResponse --> ProjectMember
    AddProjectMemberResponse --> Project
```

---

### CreateTaskRequest

<a name="createtaskrequest"></a>

CreateTaskRequest creates a new task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateTaskRequest` |
| **Field Count** | 10 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `title` | string | optional | Task title. |
| 3 | `description` | string | optional | Task description. |
| 4 | `assignee_id` | string | optional | Assignee user ID. (Must be a non-empty identifier) |
| 5 | `priority` | [`Priority`](#priority) | optional | Priority Higher values indicate higher priority. |
| 6 | `due_date` | [`Timestamp`](#timestamp) | optional | Due date. |
| 7 | `estimated_hours` | double | optional | Estimated hours. |
| 8 | `parent_task_id` | string | optional | Parent task ID (for subtasks). (Must be a non-empty identifier) |
| 9 | `dependency_ids` | string | repeated | Dependencies. |
| 10 | `labels` | string | repeated | Labels. |

#### Proto Definition

```protobuf
message CreateTaskRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // Task title.
  optional string title = 2;
  // Task description.
  optional string description = 3;
  // Assignee user ID. (Must be a non-empty identifier)
  optional string assignee_id = 4;
  // Priority Higher values indicate higher priority.
  optional Priority priority = 5;
  // Due date.
  optional Timestamp due_date = 6;
  // Estimated hours.
  optional double estimated_hours = 7;
  // Parent task ID (for subtasks). (Must be a non-empty identifier)
  optional string parent_task_id = 8;
  // Dependencies.
  repeated string dependency_ids = 9;
  // Labels.
  repeated string labels = 10;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateTaskRequest {
        +string project_id
        +string title
        +string description
        +string assignee_id
        +Priority priority
        +Timestamp due_date
        +double estimated_hours
        +string parent_task_id
        +string[] dependency_ids
        +string[] labels
    }
    CreateTaskRequest --> Priority
    CreateTaskRequest --> Timestamp
```

---

### ListTasksRequest

<a name="listtasksrequest"></a>

ListTasksRequest lists tasks.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListTasksRequest` |
| **Field Count** | 12 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `pagination` | [`PaginationRequest`](#paginationrequest) | optional | Pagination. |
| 2 | `project_id` | string | optional | Filter by project ID. (Must be a non-empty identifier) |
| 3 | `assignee_id` | string | optional | Filter by assignee ID. (Must be a non-empty identifier) |
| 4 | `status` | [`Status`](#status) | optional | Filter by status. |
| 5 | `priority` | [`Priority`](#priority) | optional | Filter by priority Higher values indicate higher priority. |
| 6 | `state` | [`ResourceState`](#resourcestate) | optional | Filter by state. |
| 7 | `labels` | string | repeated | Filter by labels. |
| 8 | `search_query` | string | optional | Search query. |
| 9 | `due_before` | [`Timestamp`](#timestamp) | optional | Due before date. |
| 10 | `due_after` | [`Timestamp`](#timestamp) | optional | Due after date. |
| 11 | `sort_by` | string | optional | Sort by field. |
| 12 | `sort_desc` | bool | optional | Sort descending. |

#### Proto Definition

```protobuf
message ListTasksRequest {
  // Pagination.
  optional PaginationRequest pagination = 1;
  // Filter by project ID. (Must be a non-empty identifier)
  optional string project_id = 2;
  // Filter by assignee ID. (Must be a non-empty identifier)
  optional string assignee_id = 3;
  // Filter by status.
  optional Status status = 4;
  // Filter by priority Higher values indicate higher priority.
  optional Priority priority = 5;
  // Filter by state.
  optional ResourceState state = 6;
  // Filter by labels.
  repeated string labels = 7;
  // Search query.
  optional string search_query = 8;
  // Due before date.
  optional Timestamp due_before = 9;
  // Due after date.
  optional Timestamp due_after = 10;
  // Sort by field.
  optional string sort_by = 11;
  // Sort descending.
  optional bool sort_desc = 12;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListTasksRequest {
        +PaginationRequest pagination
        +string project_id
        +string assignee_id
        +Status status
        +Priority priority
        +ResourceState state
        +string[] labels
        +string search_query
        +Timestamp due_before
        +Timestamp due_after
        +string sort_by
        +bool sort_desc
    }
    ListTasksRequest --> PaginationRequest
    ListTasksRequest --> Status
    ListTasksRequest --> Priority
    ListTasksRequest --> ResourceState
    ListTasksRequest --> Timestamp
    ListTasksRequest --> Timestamp
```

---

### ProjectSyncMessage

<a name="projectsyncmessage"></a>

ProjectSyncMessage for bidirectional sync.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ProjectSyncMessage` |
| **Field Count** | 8 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `message_id` | string | optional | Message ID for tracking. (Must be a non-empty identifier) |
| 2 | `message_type` | string | optional | Message type (update, acknowledge, error). |
| 3 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |
| 4 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 5 | `project` | [`Project`](#project) | optional | Updated project data. |
| 6 | `tasks` | [`Task`](#task) | repeated | Updated tasks. |
| 7 | `sync_status` | string | optional | Sync status. |
| 8 | `error` | string | optional | Error message if applicable. |

#### Proto Definition

```protobuf
message ProjectSyncMessage {
  // Message ID for tracking. (Must be a non-empty identifier)
  optional string message_id = 1;
  // Message type (update, acknowledge, error).
  optional string message_type = 2;
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 3;
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 4;
  // Updated project data.
  optional Project project = 5;
  // Updated tasks.
  repeated Task tasks = 6;
  // Sync status.
  optional string sync_status = 7;
  // Error message if applicable.
  optional string error = 8;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ProjectSyncMessage {
        +string message_id
        +string message_type
        +Timestamp timestamp
        +string project_id
        +Project project
        +Task[] tasks
        +string sync_status
        +string error
    }
    ProjectSyncMessage --> Timestamp
    ProjectSyncMessage --> Project
    ProjectSyncMessage "1" --> "*" Task
```

---

### DeleteProjectResponse

<a name="deleteprojectresponse"></a>

DeleteProjectResponse confirms deletion.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteProjectResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `deleted_at` | [`Timestamp`](#timestamp) | optional | Deletion timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message DeleteProjectResponse {
  // Success status.
  optional bool success = 1;
  // Deletion timestamp. (RFC 3339 timestamp format)
  optional Timestamp deleted_at = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeleteProjectResponse {
        +bool success
        +Timestamp deleted_at
    }
    DeleteProjectResponse --> Timestamp
```

---

### UpdateTaskResponse

<a name="updatetaskresponse"></a>

UpdateTaskResponse returns the updated task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateTaskResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task` | [`Task`](#task) | optional | Updated task. |

#### Proto Definition

```protobuf
message UpdateTaskResponse {
  // Updated task.
  optional Task task = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateTaskResponse {
        +Task task
    }
    UpdateTaskResponse --> Task
```

---

### ProjectUpdate

<a name="projectupdate"></a>

ProjectUpdate represents a project update event.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ProjectUpdate` |
| **Field Count** | 6 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `timestamp` | [`Timestamp`](#timestamp) | optional | Update timestamp. (RFC 3339 timestamp format) |
| 2 | `update_type` | string | optional | Update type. |
| 3 | `project` | [`Project`](#project) | optional | Updated project. |
| 4 | `task` | [`Task`](#task) | optional | Updated task (if applicable). |
| 5 | `description` | string | optional | Update description. |
| 6 | `updated_by` | string | optional | User who made the update. |

#### Proto Definition

```protobuf
message ProjectUpdate {
  // Update timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 1;
  // Update type.
  optional string update_type = 2;
  // Updated project.
  optional Project project = 3;
  // Updated task (if applicable).
  optional Task task = 4;
  // Update description.
  optional string description = 5;
  // User who made the update.
  optional string updated_by = 6;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ProjectUpdate {
        +Timestamp timestamp
        +string update_type
        +Project project
        +Task task
        +string description
        +string updated_by
    }
    ProjectUpdate --> Timestamp
    ProjectUpdate --> Project
    ProjectUpdate --> Task
```

---

### AssignTaskResponse

<a name="assigntaskresponse"></a>

AssignTaskResponse returns the assigned task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AssignTaskResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task` | [`Task`](#task) | optional | Updated task. |

#### Proto Definition

```protobuf
message AssignTaskResponse {
  // Updated task.
  optional Task task = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AssignTaskResponse {
        +Task task
    }
    AssignTaskResponse --> Task
```

---

### CompleteTaskRequest

<a name="completetaskrequest"></a>

CompleteTaskRequest marks a task as completed.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CompleteTaskRequest` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task_id` | string | optional | Task ID. (Must be a non-empty identifier) |
| 2 | `actual_hours` | double | optional | Actual hours spent. |
| 3 | `notes` | string | optional | Completion notes. |

#### Proto Definition

```protobuf
message CompleteTaskRequest {
  // Task ID. (Must be a non-empty identifier)
  optional string task_id = 1;
  // Actual hours spent.
  optional double actual_hours = 2;
  // Completion notes.
  optional string notes = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CompleteTaskRequest {
        +string task_id
        +double actual_hours
        +string notes
    }
```

---

### StreamProjectUpdatesRequest

<a name="streamprojectupdatesrequest"></a>

StreamProjectUpdatesRequest requests project update stream.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.StreamProjectUpdatesRequest` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `include_tasks` | bool | optional | Include task updates. |
| 3 | `include_members` | bool | optional | Include member changes. |

#### Proto Definition

```protobuf
message StreamProjectUpdatesRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // Include task updates.
  optional bool include_tasks = 2;
  // Include member changes.
  optional bool include_members = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class StreamProjectUpdatesRequest {
        +string project_id
        +bool include_tasks
        +bool include_members
    }
```

---

### RemoveProjectMemberRequest

<a name="removeprojectmemberrequest"></a>

RemoveProjectMemberRequest removes a member from a project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.RemoveProjectMemberRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `user_id` | string | optional | User ID to remove. (Must be a non-empty identifier) |

#### Proto Definition

```protobuf
message RemoveProjectMemberRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // User ID to remove. (Must be a non-empty identifier)
  optional string user_id = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RemoveProjectMemberRequest {
        +string project_id
        +string user_id
    }
```

---

### GetTaskResponse

<a name="gettaskresponse"></a>

GetTaskResponse returns the requested task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetTaskResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task` | [`Task`](#task) | optional | Retrieved task. |

#### Proto Definition

```protobuf
message GetTaskResponse {
  // Retrieved task.
  optional Task task = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetTaskResponse {
        +Task task
    }
    GetTaskResponse --> Task
```

---

### ListTasksResponse

<a name="listtasksresponse"></a>

ListTasksResponse returns matching tasks.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListTasksResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `tasks` | [`Task`](#task) | repeated | Matching tasks. |
| 2 | `pagination` | [`PaginationResponse`](#paginationresponse) | optional | Pagination metadata. |

#### Proto Definition

```protobuf
message ListTasksResponse {
  // Matching tasks.
  repeated Task tasks = 1;
  // Pagination metadata.
  optional PaginationResponse pagination = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListTasksResponse {
        +Task[] tasks
        +PaginationResponse pagination
    }
    ListTasksResponse "1" --> "*" Task
    ListTasksResponse --> PaginationResponse
```

---

### BatchCreateTasksResponse

<a name="batchcreatetasksresponse"></a>

BatchCreateTasksResponse returns batch creation results.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.BatchCreateTasksResponse` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `tasks` | [`Task`](#task) | repeated | Created tasks. |
| 2 | `created_count` | int32 | optional | Number of tasks created Must be >= 0. |
| 3 | `failed_count` | int32 | optional | Number of tasks failed Must be >= 0. |
| 4 | `errors` | [`Error`](#error) | repeated | Errors encountered. |

#### Proto Definition

```protobuf
message BatchCreateTasksResponse {
  // Created tasks.
  repeated Task tasks = 1;
  // Number of tasks created Must be >= 0.
  optional int32 created_count = 2;
  // Number of tasks failed Must be >= 0.
  optional int32 failed_count = 3;
  // Errors encountered.
  repeated Error errors = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class BatchCreateTasksResponse {
        +Task[] tasks
        +int32 created_count
        +int32 failed_count
        +Error[] errors
    }
    BatchCreateTasksResponse "1" --> "*" Task
    BatchCreateTasksResponse "1" --> "*" Error
```

---

### CreateProjectResponse

<a name="createprojectresponse"></a>

CreateProjectResponse returns the created project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateProjectResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project` | [`Project`](#project) | optional | Created project. |

#### Proto Definition

```protobuf
message CreateProjectResponse {
  // Created project.
  optional Project project = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateProjectResponse {
        +Project project
    }
    CreateProjectResponse --> Project
```

---

### GetProjectRequest

<a name="getprojectrequest"></a>

GetProjectRequest retrieves a project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetProjectRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `include_deleted` | bool | optional | Include deleted projects. |

#### Proto Definition

```protobuf
message GetProjectRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // Include deleted projects.
  optional bool include_deleted = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetProjectRequest {
        +string project_id
        +bool include_deleted
    }
```

---

### UpdateProjectRequest

<a name="updateprojectrequest"></a>

UpdateProjectRequest updates a project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateProjectRequest` |
| **Field Count** | 8 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `name` | string | optional | Updated name. |
| 3 | `description` | string | optional | Updated description. |
| 4 | `end_date` | [`Timestamp`](#timestamp) | optional | Updated end date. |
| 5 | `budget` | [`Money`](#money) | optional | Updated budget. |
| 6 | `status` | [`Status`](#status) | optional | Updated status. |
| 7 | `tags` | string | repeated | Updated tags. |
| 8 | `version` | int64 | optional | Version for optimistic locking. |

#### Proto Definition

```protobuf
message UpdateProjectRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // Updated name.
  optional string name = 2;
  // Updated description.
  optional string description = 3;
  // Updated end date.
  optional Timestamp end_date = 4;
  // Updated budget.
  optional Money budget = 5;
  // Updated status.
  optional Status status = 6;
  // Updated tags.
  repeated string tags = 7;
  // Version for optimistic locking.
  optional int64 version = 8;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateProjectRequest {
        +string project_id
        +string name
        +string description
        +Timestamp end_date
        +Money budget
        +Status status
        +string[] tags
        +int64 version
    }
    UpdateProjectRequest --> Timestamp
    UpdateProjectRequest --> Money
    UpdateProjectRequest --> Status
```

---

### UpdateProjectResponse

<a name="updateprojectresponse"></a>

UpdateProjectResponse returns the updated project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateProjectResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project` | [`Project`](#project) | optional | Updated project. |

#### Proto Definition

```protobuf
message UpdateProjectResponse {
  // Updated project.
  optional Project project = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateProjectResponse {
        +Project project
    }
    UpdateProjectResponse --> Project
```

---

### GetTaskRequest

<a name="gettaskrequest"></a>

GetTaskRequest retrieves a task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.GetTaskRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task_id` | string | optional | Task ID. (Must be a non-empty identifier) |
| 2 | `include_deleted` | bool | optional | Include deleted tasks. |

#### Proto Definition

```protobuf
message GetTaskRequest {
  // Task ID. (Must be a non-empty identifier)
  optional string task_id = 1;
  // Include deleted tasks.
  optional bool include_deleted = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetTaskRequest {
        +string task_id
        +bool include_deleted
    }
```

---

### DeleteTaskRequest

<a name="deletetaskrequest"></a>

DeleteTaskRequest deletes a task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteTaskRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task_id` | string | optional | Task ID. (Must be a non-empty identifier) |
| 2 | `hard_delete` | bool | optional | Hard delete (permanent). |

#### Proto Definition

```protobuf
message DeleteTaskRequest {
  // Task ID. (Must be a non-empty identifier)
  optional string task_id = 1;
  // Hard delete (permanent).
  optional bool hard_delete = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeleteTaskRequest {
        +string task_id
        +bool hard_delete
    }
```

---

### DeleteProjectRequest

<a name="deleteprojectrequest"></a>

DeleteProjectRequest deletes a project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteProjectRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `hard_delete` | bool | optional | Hard delete (permanent). |

#### Proto Definition

```protobuf
message DeleteProjectRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // Hard delete (permanent).
  optional bool hard_delete = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeleteProjectRequest {
        +string project_id
        +bool hard_delete
    }
```

---

### ListProjectsResponse

<a name="listprojectsresponse"></a>

ListProjectsResponse returns matching projects.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListProjectsResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `projects` | [`Project`](#project) | repeated | Matching projects. |
| 2 | `pagination` | [`PaginationResponse`](#paginationresponse) | optional | Pagination metadata. |

#### Proto Definition

```protobuf
message ListProjectsResponse {
  // Matching projects.
  repeated Project projects = 1;
  // Pagination metadata.
  optional PaginationResponse pagination = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListProjectsResponse {
        +Project[] projects
        +PaginationResponse pagination
    }
    ListProjectsResponse "1" --> "*" Project
    ListProjectsResponse --> PaginationResponse
```

---

### AddProjectMemberRequest

<a name="addprojectmemberrequest"></a>

AddProjectMemberRequest adds a member to a project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AddProjectMemberRequest` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID. (Must be a non-empty identifier) |
| 2 | `user_id` | string | optional | User ID to add. (Must be a non-empty identifier) |
| 3 | `role` | string | optional | Member role. |
| 4 | `access_level` | [`AccessLevel`](#accesslevel) | optional | Access level. |

#### Proto Definition

```protobuf
message AddProjectMemberRequest {
  // Project ID. (Must be a non-empty identifier)
  optional string project_id = 1;
  // User ID to add. (Must be a non-empty identifier)
  optional string user_id = 2;
  // Member role.
  optional string role = 3;
  // Access level.
  optional AccessLevel access_level = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AddProjectMemberRequest {
        +string project_id
        +string user_id
        +string role
        +AccessLevel access_level
    }
    AddProjectMemberRequest --> AccessLevel
```

---

### UpdateTaskRequest

<a name="updatetaskrequest"></a>

UpdateTaskRequest updates a task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.UpdateTaskRequest` |
| **Field Count** | 9 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task_id` | string | optional | Task ID. (Must be a non-empty identifier) |
| 2 | `title` | string | optional | Updated title. |
| 3 | `description` | string | optional | Updated description. |
| 4 | `assignee_id` | string | optional | Updated assignee. (Must be a non-empty identifier) |
| 5 | `priority` | [`Priority`](#priority) | optional | Updated priority Higher values indicate higher priority. |
| 6 | `status` | [`Status`](#status) | optional | Updated status. |
| 7 | `due_date` | [`Timestamp`](#timestamp) | optional | Updated due date. |
| 8 | `actual_hours` | double | optional | Actual hours. |
| 9 | `version` | int64 | optional | Version for optimistic locking. |

#### Proto Definition

```protobuf
message UpdateTaskRequest {
  // Task ID. (Must be a non-empty identifier)
  optional string task_id = 1;
  // Updated title.
  optional string title = 2;
  // Updated description.
  optional string description = 3;
  // Updated assignee. (Must be a non-empty identifier)
  optional string assignee_id = 4;
  // Updated priority Higher values indicate higher priority.
  optional Priority priority = 5;
  // Updated status.
  optional Status status = 6;
  // Updated due date.
  optional Timestamp due_date = 7;
  // Actual hours.
  optional double actual_hours = 8;
  // Version for optimistic locking.
  optional int64 version = 9;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateTaskRequest {
        +string task_id
        +string title
        +string description
        +string assignee_id
        +Priority priority
        +Status status
        +Timestamp due_date
        +double actual_hours
        +int64 version
    }
    UpdateTaskRequest --> Priority
    UpdateTaskRequest --> Status
    UpdateTaskRequest --> Timestamp
```

---

### CompleteTaskResponse

<a name="completetaskresponse"></a>

CompleteTaskResponse returns the completed task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CompleteTaskResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task` | [`Task`](#task) | optional | Completed task. |
| 2 | `completed_at` | [`Timestamp`](#timestamp) | optional | Completion timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message CompleteTaskResponse {
  // Completed task.
  optional Task task = 1;
  // Completion timestamp. (RFC 3339 timestamp format)
  optional Timestamp completed_at = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CompleteTaskResponse {
        +Task task
        +Timestamp completed_at
    }
    CompleteTaskResponse --> Task
    CompleteTaskResponse --> Timestamp
```

---

### CreateProjectRequest

<a name="createprojectrequest"></a>

CreateProjectRequest creates a new project.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateProjectRequest` |
| **Field Count** | 8 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Project name. |
| 2 | `description` | string | optional | Project description. |
| 3 | `owner_id` | string | optional | Owner user ID. (Must be a non-empty identifier) |
| 4 | `start_date` | [`Timestamp`](#timestamp) | optional | Start date. |
| 5 | `end_date` | [`Timestamp`](#timestamp) | optional | End date. |
| 6 | `budget` | [`Money`](#money) | optional | Budget. |
| 7 | `members` | [`ProjectMember`](#projectmember) | repeated | Initial members. |
| 8 | `tags` | string | repeated | Tags. |

#### Proto Definition

```protobuf
message CreateProjectRequest {
  // Project name.
  optional string name = 1;
  // Project description.
  optional string description = 2;
  // Owner user ID. (Must be a non-empty identifier)
  optional string owner_id = 3;
  // Start date.
  optional Timestamp start_date = 4;
  // End date.
  optional Timestamp end_date = 5;
  // Budget.
  optional Money budget = 6;
  // Initial members.
  repeated ProjectMember members = 7;
  // Tags.
  repeated string tags = 8;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateProjectRequest {
        +string name
        +string description
        +string owner_id
        +Timestamp start_date
        +Timestamp end_date
        +Money budget
        +ProjectMember[] members
        +string[] tags
    }
    CreateProjectRequest --> Timestamp
    CreateProjectRequest --> Timestamp
    CreateProjectRequest --> Money
    CreateProjectRequest "1" --> "*" ProjectMember
```

---

### RemoveProjectMemberResponse

<a name="removeprojectmemberresponse"></a>

RemoveProjectMemberResponse confirms removal.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.RemoveProjectMemberResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `project` | [`Project`](#project) | optional | Updated project. |

#### Proto Definition

```protobuf
message RemoveProjectMemberResponse {
  // Success status.
  optional bool success = 1;
  // Updated project.
  optional Project project = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RemoveProjectMemberResponse {
        +bool success
        +Project project
    }
    RemoveProjectMemberResponse --> Project
```

---

### ListProjectsRequest

<a name="listprojectsrequest"></a>

ListProjectsRequest lists projects.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.ListProjectsRequest` |
| **Field Count** | 8 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `pagination` | [`PaginationRequest`](#paginationrequest) | optional | Pagination. |
| 2 | `owner_id` | string | optional | Filter by owner ID. (Must be a non-empty identifier) |
| 3 | `state` | [`ResourceState`](#resourcestate) | optional | Filter by state. |
| 4 | `status` | [`Status`](#status) | optional | Filter by status. |
| 5 | `tags` | string | repeated | Filter by tags. |
| 6 | `search_query` | string | optional | Search query. |
| 7 | `sort_by` | string | optional | Sort by field. |
| 8 | `sort_desc` | bool | optional | Sort descending. |

#### Proto Definition

```protobuf
message ListProjectsRequest {
  // Pagination.
  optional PaginationRequest pagination = 1;
  // Filter by owner ID. (Must be a non-empty identifier)
  optional string owner_id = 2;
  // Filter by state.
  optional ResourceState state = 3;
  // Filter by status.
  optional Status status = 4;
  // Filter by tags.
  repeated string tags = 5;
  // Search query.
  optional string search_query = 6;
  // Sort by field.
  optional string sort_by = 7;
  // Sort descending.
  optional bool sort_desc = 8;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListProjectsRequest {
        +PaginationRequest pagination
        +string owner_id
        +ResourceState state
        +Status status
        +string[] tags
        +string search_query
        +string sort_by
        +bool sort_desc
    }
    ListProjectsRequest --> PaginationRequest
    ListProjectsRequest --> ResourceState
    ListProjectsRequest --> Status
```

---

### CreateTaskResponse

<a name="createtaskresponse"></a>

CreateTaskResponse returns the created task.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.CreateTaskResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task` | [`Task`](#task) | optional | Created task. |

#### Proto Definition

```protobuf
message CreateTaskResponse {
  // Created task.
  optional Task task = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateTaskResponse {
        +Task task
    }
    CreateTaskResponse --> Task
```

---

### DeleteTaskResponse

<a name="deletetaskresponse"></a>

DeleteTaskResponse confirms deletion.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.DeleteTaskResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `deleted_at` | [`Timestamp`](#timestamp) | optional | Deletion timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message DeleteTaskResponse {
  // Success status.
  optional bool success = 1;
  // Deletion timestamp. (RFC 3339 timestamp format)
  optional Timestamp deleted_at = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeleteTaskResponse {
        +bool success
        +Timestamp deleted_at
    }
    DeleteTaskResponse --> Timestamp
```

---

### AssignTaskRequest

<a name="assigntaskrequest"></a>

AssignTaskRequest assigns a task to a user.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `first.v1.AssignTaskRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `task_id` | string | optional | Task ID. (Must be a non-empty identifier) |
| 2 | `assignee_id` | string | optional | Assignee user ID. (Must be a non-empty identifier) |

#### Proto Definition

```protobuf
message AssignTaskRequest {
  // Task ID. (Must be a non-empty identifier)
  optional string task_id = 1;
  // Assignee user ID. (Must be a non-empty identifier)
  optional string assignee_id = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AssignTaskRequest {
        +string task_id
        +string assignee_id
    }
```

---

## 🗄️ Data Model (ERD)

<a name="erd"></a>

Entity-Relationship diagram showing the data model.

```mermaid
%{init: {'theme':'forest'}}%
erDiagram
    GetProjectResponse {
        Project project
    }

    GetProjectResponse ||--|| Project : has
    AddProjectMemberResponse {
        ProjectMember member
        Project project
    }

    AddProjectMemberResponse ||--|| ProjectMember : has
    AddProjectMemberResponse ||--|| Project : has
    CreateTaskRequest {
        string project_id
        string title
        string description
        string assignee_id
        Priority priority
        Timestamp due_date
        double estimated_hours
        string parent_task_id
        string dependency_ids
        string labels
    }

    CreateTaskRequest ||--|| Priority : has
    ListTasksRequest {
        PaginationRequest pagination
        string project_id
        string assignee_id
        Status status
        Priority priority
        ResourceState state
        string labels
        string search_query
        Timestamp due_before
        Timestamp due_after
        string sort_by
        bool sort_desc
    }

    ListTasksRequest ||--|| PaginationRequest : has
    ListTasksRequest ||--|| Status : has
    ListTasksRequest ||--|| Priority : has
    ListTasksRequest ||--|| ResourceState : has
    ProjectSyncMessage {
        string message_id
        string message_type
        Timestamp timestamp
        string project_id
        Project project
        Task tasks
        string sync_status
        string error
    }

    ProjectSyncMessage ||--|| Project : has
    ProjectSyncMessage ||--o{ Task : has
    DeleteProjectResponse {
        bool success
        Timestamp deleted_at
    }

    UpdateTaskResponse {
        Task task
    }

    UpdateTaskResponse ||--|| Task : has
    ProjectUpdate {
        Timestamp timestamp
        string update_type
        Project project
        Task task
        string description
        string updated_by
    }

    ProjectUpdate ||--|| Project : has
    ProjectUpdate ||--|| Task : has
    AssignTaskResponse {
        Task task
    }

    AssignTaskResponse ||--|| Task : has
    CompleteTaskRequest {
        string task_id
        double actual_hours
        string notes
    }

    StreamProjectUpdatesRequest {
        string project_id
        bool include_tasks
        bool include_members
    }

    RemoveProjectMemberRequest {
        string project_id
        string user_id
    }

    GetTaskResponse {
        Task task
    }

    GetTaskResponse ||--|| Task : has
    ListTasksResponse {
        Task tasks
        PaginationResponse pagination
    }

    ListTasksResponse ||--o{ Task : has
    ListTasksResponse ||--|| PaginationResponse : has
    BatchCreateTasksResponse {
        Task tasks
        int32 created_count
        int32 failed_count
        Error errors
    }

    BatchCreateTasksResponse ||--o{ Task : has
    BatchCreateTasksResponse ||--o{ Error : has
    CreateProjectResponse {
        Project project
    }

    CreateProjectResponse ||--|| Project : has
    GetProjectRequest {
        string project_id
        bool include_deleted
    }

    UpdateProjectRequest {
        string project_id
        string name
        string description
        Timestamp end_date
        Money budget
        Status status
        string tags
        int64 version
    }

    UpdateProjectRequest ||--|| Money : has
    UpdateProjectRequest ||--|| Status : has
    UpdateProjectResponse {
        Project project
    }

    UpdateProjectResponse ||--|| Project : has
    GetTaskRequest {
        string task_id
        bool include_deleted
    }

    DeleteTaskRequest {
        string task_id
        bool hard_delete
    }

    DeleteProjectRequest {
        string project_id
        bool hard_delete
    }

    ListProjectsResponse {
        Project projects
        PaginationResponse pagination
    }

    ListProjectsResponse ||--o{ Project : has
    ListProjectsResponse ||--|| PaginationResponse : has
    AddProjectMemberRequest {
        string project_id
        string user_id
        string role
        AccessLevel access_level
    }

    AddProjectMemberRequest ||--|| AccessLevel : has
    UpdateTaskRequest {
        string task_id
        string title
        string description
        string assignee_id
        Priority priority
        Status status
        Timestamp due_date
        double actual_hours
        int64 version
    }

    UpdateTaskRequest ||--|| Priority : has
    UpdateTaskRequest ||--|| Status : has
    CompleteTaskResponse {
        Task task
        Timestamp completed_at
    }

    CompleteTaskResponse ||--|| Task : has
    CreateProjectRequest {
        string name
        string description
        string owner_id
        Timestamp start_date
        Timestamp end_date
        Money budget
        ProjectMember members
        string tags
    }

    CreateProjectRequest ||--|| Money : has
    CreateProjectRequest ||--o{ ProjectMember : has
    RemoveProjectMemberResponse {
        bool success
        Project project
    }

    RemoveProjectMemberResponse ||--|| Project : has
    ListProjectsRequest {
        PaginationRequest pagination
        string owner_id
        ResourceState state
        Status status
        string tags
        string search_query
        string sort_by
        bool sort_desc
    }

    ListProjectsRequest ||--|| PaginationRequest : has
    ListProjectsRequest ||--|| ResourceState : has
    ListProjectsRequest ||--|| Status : has
    CreateTaskResponse {
        Task task
    }

    CreateTaskResponse ||--|| Task : has
    DeleteTaskResponse {
        bool success
        Timestamp deleted_at
    }

    AssignTaskRequest {
        string task_id
        string assignee_id
    }

```

---

## ⚠️ Error Codes

<a name="error-codes"></a>

This service uses standard gRPC status codes:

| gRPC Code | HTTP Status | Description |
|-----------|-------------|-------------|
| `OK` | 200 | Success |
| `CANCELLED` | 499 | Operation cancelled by client |
| `UNKNOWN` | 500 | Unknown error |
| `INVALID_ARGUMENT` | 400 | Client specified an invalid argument |
| `DEADLINE_EXCEEDED` | 504 | Deadline expired before operation could complete |
| `NOT_FOUND` | 404 | Requested entity not found |
| `ALREADY_EXISTS` | 409 | Entity already exists |
| `PERMISSION_DENIED` | 403 | Caller does not have permission |
| `RESOURCE_EXHAUSTED` | 429 | Resource has been exhausted |
| `FAILED_PRECONDITION` | 400 | Operation rejected because system is not in required state |
| `ABORTED` | 409 | Operation aborted, typically due to concurrency issue |
| `OUT_OF_RANGE` | 400 | Operation attempted past valid range |
| `UNIMPLEMENTED` | 501 | Operation not implemented |
| `INTERNAL` | 500 | Internal server error |
| `UNAVAILABLE` | 503 | Service unavailable |
| `DATA_LOSS` | 500 | Unrecoverable data loss or corruption |
| `UNAUTHENTICATED` | 401 | Request does not have valid authentication credentials |

### Error Handling Best Practices

1. **Check status codes**: Always check the gRPC status code before processing responses
2. **Implement retries**: Use exponential backoff for transient errors (`UNAVAILABLE`, `RESOURCE_EXHAUSTED`)
3. **Log errors**: Log error details with correlation IDs for troubleshooting
4. **Handle streaming errors**: Properly handle errors in streaming RPCs
5. **Validate inputs**: Validate inputs client-side to avoid `INVALID_ARGUMENT` errors

---

## 💡 Examples

<a name="examples"></a>

### Go Example

```go
package main

import (
    "context"
    "log"
    "time"

    "google.golang.org/grpc"
    "google.golang.org/grpc/credentials/insecure"

    pb "first.v1"
)

func main() {
    // Connect to the service
    conn, err := grpc.Dial("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewFirstServiceClient(conn)

    // Example RPC call
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &pb.CreateProjectRequest{
        // Fill in request fields
    }

    resp, err := client.CreateProject(ctx, req)
    if err != nil {
        log.Fatalf("RPC failed: %v", err)
    }

    log.Printf("Response: %v", resp)
}
```

### TypeScript Example

```typescript
import * as grpc from '@grpc/grpc-js';
import * as protoLoader from '@grpc/proto-loader';
import { ProtoGrpcType } from './first/first';
import { FirstServiceClient } from './first.v1/FirstService';

// Load proto file
const packageDefinition = protoLoader.loadSync(
    'first/first.proto',
    {
        keepCase: true,
        longs: String,
        enums: String,
        defaults: true,
        oneofs: true
    }
);

const proto = grpc.loadPackageDefinition(
    packageDefinition
) as unknown as ProtoGrpcType;

// Create client
const client: FirstServiceClient = new proto.first.v1.FirstService(
    'localhost:50051',
    grpc.credentials.createInsecure()
);

// Example RPC call
const request = {
    // Fill in request fields
};

client.CreateProject(request, (error: grpc.ServiceError | null, response?: any) => {
    if (error) {
        console.error('RPC failed:', error);
        return;
    }
    console.log('Response:', response);
});
```

---

---

<div align="center">

**Generated Documentation**

| Attribute | Value |
|-----------|-------|
| Generated At | 2025-11-23 01:12:47 UTC |
| Generator Version | 7.0.0 |

📚 **Documentation** | 🔧 **ProtoDocs** | ✨ **Auto-Generated**

</div>
