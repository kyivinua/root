# 📚 SecondService API Documentation

| **Attribute** | **Value** |
|---------------|----------|
| **Service Name** | `SecondService` |
| **Package** | `second.v1` |
| **Version** | v1 |
| **Proto File** | `second/second.proto` |
| **Generated** | 2025-11-23T00:35:25Z |

SecondService manages documents, permissions, and resource activities.

---

## 📑 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [gRPC Service Interactions](#service-interaction)
- [Message Type Diagrams](#class-diagram)
- [Methods](#methods)
  - [CreateDocument](#createdocument)
  - [GetDocument](#getdocument)
  - [UpdateDocument](#updatedocument)
  - [DeleteDocument](#deletedocument)
  - [ListDocuments](#listdocuments)
  - [AddCollaborator](#addcollaborator)
  - [RemoveCollaborator](#removecollaborator)
  - [GrantPermission](#grantpermission)
  - [RevokePermission](#revokepermission)
  - [CheckPermission](#checkpermission)
  - [ListPermissions](#listpermissions)
  - [LogActivity](#logactivity)
  - [GetActivityLog](#getactivitylog)
  - [GetQuota](#getquota)
  - [UpdateQuota](#updatequota)
  - [StreamActivityFeed](#streamactivityfeed)
  - [BatchGrantPermissions](#batchgrantpermissions)
  - [CollaborateOnDocument](#collaborateondocument)
- [Messages](#messages)
  - [AddCollaboratorRequest](#addcollaboratorrequest)
  - [GetQuotaResponse](#getquotaresponse)
  - [BatchGrantPermissionsResponse](#batchgrantpermissionsresponse)
  - [DocumentCollaboration](#documentcollaboration)
  - [CreateDocumentRequest](#createdocumentrequest)
  - [UpdateDocumentResponse](#updatedocumentresponse)
  - [RemoveCollaboratorRequest](#removecollaboratorrequest)
  - [ListPermissionsRequest](#listpermissionsrequest)
  - [ListPermissionsResponse](#listpermissionsresponse)
  - [GetDocumentResponse](#getdocumentresponse)
  - [DeleteDocumentRequest](#deletedocumentrequest)
  - [GrantPermissionRequest](#grantpermissionrequest)
  - [LogActivityRequest](#logactivityrequest)
  - [CursorPosition](#cursorposition)
  - [GetDocumentRequest](#getdocumentrequest)
  - [ListDocumentsRequest](#listdocumentsrequest)
  - [ListDocumentsResponse](#listdocumentsresponse)
  - [LogActivityResponse](#logactivityresponse)
  - [CheckPermissionRequest](#checkpermissionrequest)
  - [CheckPermissionResponse](#checkpermissionresponse)
  - [CreateDocumentResponse](#createdocumentresponse)
  - [GrantPermissionResponse](#grantpermissionresponse)
  - [StreamActivityFeedRequest](#streamactivityfeedrequest)
  - [GetQuotaRequest](#getquotarequest)
  - [UpdateDocumentRequest](#updatedocumentrequest)
  - [DeleteDocumentResponse](#deletedocumentresponse)
  - [AddCollaboratorResponse](#addcollaboratorresponse)
  - [RevokePermissionResponse](#revokepermissionresponse)
  - [GetActivityLogResponse](#getactivitylogresponse)
  - [UpdateQuotaRequest](#updatequotarequest)
  - [RemoveCollaboratorResponse](#removecollaboratorresponse)
  - [RevokePermissionRequest](#revokepermissionrequest)
  - [GetActivityLogRequest](#getactivitylogrequest)
  - [UpdateQuotaResponse](#updatequotaresponse)
  - [ContentChange](#contentchange)
  - [Position](#position)
- [Error Codes](#error-codes)
- [Examples](#examples)

---

## 📖 Overview

<a name="overview"></a>

### Service Statistics

| Metric | Count |
|--------|-------|
| **RPC Methods** | 18 |
| **Message Types** | 36 |
| **Enumerations** | 0 |
| **Streaming RPCs** | 3 |

### Quick Start

This service provides the following capabilities:

- [`CreateDocument`](#createdocument): CreateDocument creates a new document.
- [`GetDocument`](#getdocument): GetDocument retrieves a document by ID.
- [`UpdateDocument`](#updatedocument): UpdateDocument updates an existing document.
- [`DeleteDocument`](#deletedocument): DeleteDocument soft-deletes a document.
- [`ListDocuments`](#listdocuments): ListDocuments lists documents with pagination and filtering.
- ... and 13 more methods

---

## 🏗️ Architecture

<a name="architecture"></a>

```mermaid
%{init: {'theme':'forest'}}%
graph TB
    classDef serviceClass fill:#e1f5ff,stroke:#01579b,stroke-width:3px
    classDef methodClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef messageClass fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px

    SecondService[🔧 SecondService]:::serviceClass

    CreateDocument[CreateDocument]:::methodClass
    SecondService --> CreateDocument
    CreateDocument_in[📥 CreateDocumentRequest]:::messageClass
    CreateDocument_out[📤 CreateDocumentResponse]:::messageClass
    CreateDocument_in -.->|input| CreateDocument
    CreateDocument -.->|output| CreateDocument_out
    GetDocument[GetDocument]:::methodClass
    SecondService --> GetDocument
    GetDocument_in[📥 GetDocumentRequest]:::messageClass
    GetDocument_out[📤 GetDocumentResponse]:::messageClass
    GetDocument_in -.->|input| GetDocument
    GetDocument -.->|output| GetDocument_out
    UpdateDocument[UpdateDocument]:::methodClass
    SecondService --> UpdateDocument
    UpdateDocument_in[📥 UpdateDocumentRequest]:::messageClass
    UpdateDocument_out[📤 UpdateDocumentResponse]:::messageClass
    UpdateDocument_in -.->|input| UpdateDocument
    UpdateDocument -.->|output| UpdateDocument_out
    DeleteDocument[DeleteDocument]:::methodClass
    SecondService --> DeleteDocument
    DeleteDocument_in[📥 DeleteDocumentRequest]:::messageClass
    DeleteDocument_out[📤 DeleteDocumentResponse]:::messageClass
    DeleteDocument_in -.->|input| DeleteDocument
    DeleteDocument -.->|output| DeleteDocument_out
    ListDocuments[ListDocuments]:::methodClass
    SecondService --> ListDocuments
    ListDocuments_in[📥 ListDocumentsRequest]:::messageClass
    ListDocuments_out[📤 ListDocumentsResponse]:::messageClass
    ListDocuments_in -.->|input| ListDocuments
    ListDocuments -.->|output| ListDocuments_out
    AddCollaborator[AddCollaborator]:::methodClass
    SecondService --> AddCollaborator
    AddCollaborator_in[📥 AddCollaboratorRequest]:::messageClass
    AddCollaborator_out[📤 AddCollaboratorResponse]:::messageClass
    AddCollaborator_in -.->|input| AddCollaborator
    AddCollaborator -.->|output| AddCollaborator_out
    RemoveCollaborator[RemoveCollaborator]:::methodClass
    SecondService --> RemoveCollaborator
    RemoveCollaborator_in[📥 RemoveCollaboratorRequest]:::messageClass
    RemoveCollaborator_out[📤 RemoveCollaboratorResponse]:::messageClass
    RemoveCollaborator_in -.->|input| RemoveCollaborator
    RemoveCollaborator -.->|output| RemoveCollaborator_out
    GrantPermission[GrantPermission]:::methodClass
    SecondService --> GrantPermission
    GrantPermission_in[📥 GrantPermissionRequest]:::messageClass
    GrantPermission_out[📤 GrantPermissionResponse]:::messageClass
    GrantPermission_in -.->|input| GrantPermission
    GrantPermission -.->|output| GrantPermission_out
    RevokePermission[RevokePermission]:::methodClass
    SecondService --> RevokePermission
    RevokePermission_in[📥 RevokePermissionRequest]:::messageClass
    RevokePermission_out[📤 RevokePermissionResponse]:::messageClass
    RevokePermission_in -.->|input| RevokePermission
    RevokePermission -.->|output| RevokePermission_out
    CheckPermission[CheckPermission]:::methodClass
    SecondService --> CheckPermission
    CheckPermission_in[📥 CheckPermissionRequest]:::messageClass
    CheckPermission_out[📤 CheckPermissionResponse]:::messageClass
    CheckPermission_in -.->|input| CheckPermission
    CheckPermission -.->|output| CheckPermission_out
    ListPermissions[ListPermissions]:::methodClass
    SecondService --> ListPermissions
    ListPermissions_in[📥 ListPermissionsRequest]:::messageClass
    ListPermissions_out[📤 ListPermissionsResponse]:::messageClass
    ListPermissions_in -.->|input| ListPermissions
    ListPermissions -.->|output| ListPermissions_out
    LogActivity[LogActivity]:::methodClass
    SecondService --> LogActivity
    LogActivity_in[📥 LogActivityRequest]:::messageClass
    LogActivity_out[📤 LogActivityResponse]:::messageClass
    LogActivity_in -.->|input| LogActivity
    LogActivity -.->|output| LogActivity_out
    GetActivityLog[GetActivityLog]:::methodClass
    SecondService --> GetActivityLog
    GetActivityLog_in[📥 GetActivityLogRequest]:::messageClass
    GetActivityLog_out[📤 GetActivityLogResponse]:::messageClass
    GetActivityLog_in -.->|input| GetActivityLog
    GetActivityLog -.->|output| GetActivityLog_out
    GetQuota[GetQuota]:::methodClass
    SecondService --> GetQuota
    GetQuota_in[📥 GetQuotaRequest]:::messageClass
    GetQuota_out[📤 GetQuotaResponse]:::messageClass
    GetQuota_in -.->|input| GetQuota
    GetQuota -.->|output| GetQuota_out
    UpdateQuota[UpdateQuota]:::methodClass
    SecondService --> UpdateQuota
    UpdateQuota_in[📥 UpdateQuotaRequest]:::messageClass
    UpdateQuota_out[📤 UpdateQuotaResponse]:::messageClass
    UpdateQuota_in -.->|input| UpdateQuota
    UpdateQuota -.->|output| UpdateQuota_out
    StreamActivityFeed[↓ StreamActivityFeed]:::methodClass
    SecondService --> StreamActivityFeed
    StreamActivityFeed_in[📥 StreamActivityFeedRequest]:::messageClass
    StreamActivityFeed_out[📤 ResourceActivity]:::messageClass
    StreamActivityFeed_in -.->|input| StreamActivityFeed
    StreamActivityFeed -.->|output| StreamActivityFeed_out
    BatchGrantPermissions[↑ BatchGrantPermissions]:::methodClass
    SecondService --> BatchGrantPermissions
    BatchGrantPermissions_in[📥 GrantPermissionRequest]:::messageClass
    BatchGrantPermissions_out[📤 BatchGrantPermissionsResponse]:::messageClass
    BatchGrantPermissions_in -.->|input| BatchGrantPermissions
    BatchGrantPermissions -.->|output| BatchGrantPermissions_out
    CollaborateOnDocument[↔️ CollaborateOnDocument]:::methodClass
    SecondService --> CollaborateOnDocument
    CollaborateOnDocument_in[📥 DocumentCollaboration]:::messageClass
    CollaborateOnDocument_out[📤 DocumentCollaboration]:::messageClass
    CollaborateOnDocument_in -.->|input| CollaborateOnDocument
    CollaborateOnDocument -.->|output| CollaborateOnDocument_out
```

---

## 🔄 gRPC Service Interactions

<a name="service-interaction"></a>

This diagram shows the interactions between the service methods and message types.

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant SecondService
    Client->>+SecondService: CreateDocument
    Note right of SecondService: CreateDocumentRequest
    SecondService->>-Client: CreateDocumentResponse
    Client->>+SecondService: GetDocument
    Note right of SecondService: GetDocumentRequest
    SecondService->>-Client: GetDocumentResponse
    Client->>+SecondService: UpdateDocument
    Note right of SecondService: UpdateDocumentRequest
    SecondService->>-Client: UpdateDocumentResponse
    Client->>+SecondService: DeleteDocument
    Note right of SecondService: DeleteDocumentRequest
    SecondService->>-Client: DeleteDocumentResponse
    Client->>+SecondService: ListDocuments
    Note right of SecondService: ListDocumentsRequest
    SecondService->>-Client: ListDocumentsResponse
    Client->>+SecondService: AddCollaborator
    Note right of SecondService: AddCollaboratorRequest
    SecondService->>-Client: AddCollaboratorResponse
    Client->>+SecondService: RemoveCollaborator
    Note right of SecondService: RemoveCollaboratorRequest
    SecondService->>-Client: RemoveCollaboratorResponse
    Client->>+SecondService: GrantPermission
    Note right of SecondService: GrantPermissionRequest
    SecondService->>-Client: GrantPermissionResponse
    Client->>+SecondService: RevokePermission
    Note right of SecondService: RevokePermissionRequest
    SecondService->>-Client: RevokePermissionResponse
    Client->>+SecondService: CheckPermission
    Note right of SecondService: CheckPermissionRequest
    SecondService->>-Client: CheckPermissionResponse
    Client->>+SecondService: ListPermissions
    Note right of SecondService: ListPermissionsRequest
    SecondService->>-Client: ListPermissionsResponse
    Client->>+SecondService: LogActivity
    Note right of SecondService: LogActivityRequest
    SecondService->>-Client: LogActivityResponse
    Client->>+SecondService: GetActivityLog
    Note right of SecondService: GetActivityLogRequest
    SecondService->>-Client: GetActivityLogResponse
    Client->>+SecondService: GetQuota
    Note right of SecondService: GetQuotaRequest
    SecondService->>-Client: GetQuotaResponse
    Client->>+SecondService: UpdateQuota
    Note right of SecondService: UpdateQuotaRequest
    SecondService->>-Client: UpdateQuotaResponse
    Client->>+SecondService: StreamActivityFeed
    Note right of SecondService: StreamActivityFeedRequest
    SecondService->>-Client: Stream of ResourceActivity
    Client->>+SecondService: BatchGrantPermissions (client stream)
    Note over Client,SecondService: Stream of GrantPermissionRequest
    SecondService->>-Client: BatchGrantPermissionsResponse
    Client->>+SecondService: CollaborateOnDocument (bidirectional stream)
    Note over Client,SecondService: Stream of DocumentCollaboration
    SecondService->>-Client: Stream of DocumentCollaboration
```

---

## 📦 Message Type Diagrams

<a name="class-diagram"></a>

UML class diagrams showing the structure of message types.

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class SecondService {
        <<service>>
        +AddCollaboratorRequest()
        +GetQuotaResponse()
        +BatchGrantPermissionsResponse()
        +DocumentCollaboration()
        +CreateDocumentRequest()
        +UpdateDocumentResponse()
        +RemoveCollaboratorRequest()
        +ListPermissionsRequest()
        +ListPermissionsResponse()
        +GetDocumentResponse()
        +DeleteDocumentRequest()
        +GrantPermissionRequest()
        +LogActivityRequest()
        +CursorPosition()
        +GetDocumentRequest()
        +ListDocumentsRequest()
        +ListDocumentsResponse()
        +LogActivityResponse()
        +CheckPermissionRequest()
        +CheckPermissionResponse()
        +CreateDocumentResponse()
        +GrantPermissionResponse()
        +StreamActivityFeedRequest()
        +GetQuotaRequest()
        +UpdateDocumentRequest()
        +DeleteDocumentResponse()
        +AddCollaboratorResponse()
        +RevokePermissionResponse()
        +GetActivityLogResponse()
        +UpdateQuotaRequest()
        +RemoveCollaboratorResponse()
        +RevokePermissionRequest()
        +GetActivityLogRequest()
        +UpdateQuotaResponse()
        +ContentChange()
        +Position()
    }

    class AddCollaboratorRequest {
        +string document_id
        +string user_id
        +AccessLevel access_level
    }

    AddCollaboratorRequest "1" --> "1" AccessLevel
    class GetQuotaResponse {
        +ResourceQuota quota
        +double usage_percentage
        +bool over_quota
    }

    GetQuotaResponse "1" --> "1" ResourceQuota
    class BatchGrantPermissionsResponse {
        +ResourcePermission permissions[]
        +int32 granted_count
        +int32 failed_count
        +Error errors[]
    }

    BatchGrantPermissionsResponse "1" --> "*" ResourcePermission
    BatchGrantPermissionsResponse "1" --> "*" Error
    class DocumentCollaboration {
        +string message_id
        +string message_type
        +Timestamp timestamp
        +string document_id
        +string user_id
        +ContentChange content_change
        +CursorPosition cursor_position
        +string comment
        +string ack_message_id
    }

    DocumentCollaboration "1" --> "1" ContentChange
    DocumentCollaboration "1" --> "1" CursorPosition
    class CreateDocumentRequest {
        +string project_id
        +string title
        +string content
        +string format
        +string author_id
        +DocumentCollaborator collaborators[]
        +string tags[]
    }

    CreateDocumentRequest "1" --> "*" DocumentCollaborator
    class UpdateDocumentResponse {
        +Document document
        +int32 new_version
    }

    UpdateDocumentResponse "1" --> "1" Document
    class RemoveCollaboratorRequest {
        +string document_id
        +string user_id
    }

    class ListPermissionsRequest {
        +string resource_id
        +ResourceType resource_type
        +string principal_id
        +bool include_expired
        +PaginationRequest pagination
    }

    ListPermissionsRequest "1" --> "1" ResourceType
    ListPermissionsRequest "1" --> "1" PaginationRequest
    class ListPermissionsResponse {
        +ResourcePermission permissions[]
        +PaginationResponse pagination
    }

    ListPermissionsResponse "1" --> "*" ResourcePermission
    ListPermissionsResponse "1" --> "1" PaginationResponse
    class GetDocumentResponse {
        +Document document
    }

    GetDocumentResponse "1" --> "1" Document
    class DeleteDocumentRequest {
        +string document_id
        +bool hard_delete
    }

    class GrantPermissionRequest {
        +string resource_id
        +ResourceType resource_type
        +string principal_id
        +string principal_type
        +AccessLevel access_level
        +Timestamp expires_at
        +string granted_by
    }

    GrantPermissionRequest "1" --> "1" ResourceType
    GrantPermissionRequest "1" --> "1" AccessLevel
    class LogActivityRequest {
        +string resource_id
        +ResourceType resource_type
        +string activity_type
        +string user_id
        +string description
        +string details
        +string ip_address
    }

    LogActivityRequest "1" --> "1" ResourceType
    class CursorPosition {
        +int32 line
        +int32 column
        +Position selection_start
        +Position selection_end
    }

    CursorPosition "1" --> "1" Position
    CursorPosition "1" --> "1" Position
    class GetDocumentRequest {
        +string document_id
        +bool include_deleted
        +bool include_content
    }

    class ListDocumentsRequest {
        +PaginationRequest pagination
        +string project_id
        +string author_id
        +ResourceState state
        +string format
        +string tags[]
        +string search_query
        +Timestamp created_after
        +Timestamp created_before
        +string sort_by
        +bool sort_desc
    }

    ListDocumentsRequest "1" --> "1" PaginationRequest
    ListDocumentsRequest "1" --> "1" ResourceState
    class ListDocumentsResponse {
        +Document documents[]
        +PaginationResponse pagination
    }

    ListDocumentsResponse "1" --> "*" Document
    ListDocumentsResponse "1" --> "1" PaginationResponse
    class LogActivityResponse {
        +ResourceActivity activity
    }

    LogActivityResponse "1" --> "1" ResourceActivity
    class CheckPermissionRequest {
        +string resource_id
        +ResourceType resource_type
        +string user_id
        +AccessLevel required_access_level
    }

    CheckPermissionRequest "1" --> "1" ResourceType
    CheckPermissionRequest "1" --> "1" AccessLevel
    class CheckPermissionResponse {
        +bool has_permission
        +AccessLevel access_level
        +ResourcePermission permission
    }

    CheckPermissionResponse "1" --> "1" AccessLevel
    CheckPermissionResponse "1" --> "1" ResourcePermission
    class CreateDocumentResponse {
        +Document document
    }

    CreateDocumentResponse "1" --> "1" Document
    class GrantPermissionResponse {
        +ResourcePermission permission
    }

    GrantPermissionResponse "1" --> "1" ResourcePermission
    class StreamActivityFeedRequest {
        +ResourceType resource_type
        +string user_id
        +string activity_types[]
        +Timestamp start_from
    }

    StreamActivityFeedRequest "1" --> "1" ResourceType
    class GetQuotaRequest {
        +string owner_id
        +ResourceType resource_type
    }

    GetQuotaRequest "1" --> "1" ResourceType
    class UpdateDocumentRequest {
        +string document_id
        +string title
        +string content
        +string format
        +int64 version
        +string update_description
    }

    class DeleteDocumentResponse {
        +bool success
        +Timestamp deleted_at
    }

    class AddCollaboratorResponse {
        +DocumentCollaborator collaborator
        +Document document
    }

    AddCollaboratorResponse "1" --> "1" DocumentCollaborator
    AddCollaboratorResponse "1" --> "1" Document
    class RevokePermissionResponse {
        +bool success
        +Timestamp revoked_at
    }

    class GetActivityLogResponse {
        +ResourceActivity activities[]
        +PaginationResponse pagination
    }

    GetActivityLogResponse "1" --> "*" ResourceActivity
    GetActivityLogResponse "1" --> "1" PaginationResponse
    class UpdateQuotaRequest {
        +string owner_id
        +ResourceType resource_type
        +int64 max_count
        +int64 max_storage_bytes
        +string updated_by
    }

    UpdateQuotaRequest "1" --> "1" ResourceType
    class RemoveCollaboratorResponse {
        +bool success
        +Document document
    }

    RemoveCollaboratorResponse "1" --> "1" Document
    class RevokePermissionRequest {
        +string permission_id
        +string revoked_by
    }

    class GetActivityLogRequest {
        +string resource_id
        +ResourceType resource_type
        +string user_id
        +string activity_type
        +Timestamp after
        +Timestamp before
        +PaginationRequest pagination
    }

    GetActivityLogRequest "1" --> "1" ResourceType
    GetActivityLogRequest "1" --> "1" PaginationRequest
    class UpdateQuotaResponse {
        +ResourceQuota quota
    }

    UpdateQuotaResponse "1" --> "1" ResourceQuota
    class ContentChange {
        +string operation
        +int32 start_position
        +int32 end_position
        +string content
        +int32 version
    }

    class Position {
        +int32 line
        +int32 column
    }

```

---

## ⚙️ Methods

<a name="methods"></a>

This service defines **18 RPC methods**:

### CreateDocument

<a name="createdocument"></a>

CreateDocument creates a new document.

#### Method Signature

```protobuf
// Unary RPC
rpc CreateDocument(CreateDocumentRequest) returns (CreateDocumentResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CreateDocument` |
| **Input Type** | [`CreateDocumentRequest`](#createdocumentrequest) |
| **Output Type** | [`CreateDocumentResponse`](#createdocumentresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CreateDocument
    Note right of Service: CreateDocumentRequest
    Service-->>-Client: Response
    Note left of Client: CreateDocumentResponse
```

---

### GetDocument

<a name="getdocument"></a>

GetDocument retrieves a document by ID.

#### Method Signature

```protobuf
// Unary RPC
rpc GetDocument(GetDocumentRequest) returns (GetDocumentResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetDocument` |
| **Input Type** | [`GetDocumentRequest`](#getdocumentrequest) |
| **Output Type** | [`GetDocumentResponse`](#getdocumentresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetDocument
    Note right of Service: GetDocumentRequest
    Service-->>-Client: Response
    Note left of Client: GetDocumentResponse
```

---

### UpdateDocument

<a name="updatedocument"></a>

UpdateDocument updates an existing document.

#### Method Signature

```protobuf
// Unary RPC
rpc UpdateDocument(UpdateDocumentRequest) returns (UpdateDocumentResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateDocument` |
| **Input Type** | [`UpdateDocumentRequest`](#updatedocumentrequest) |
| **Output Type** | [`UpdateDocumentResponse`](#updatedocumentresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: UpdateDocument
    Note right of Service: UpdateDocumentRequest
    Service-->>-Client: Response
    Note left of Client: UpdateDocumentResponse
```

---

### DeleteDocument

<a name="deletedocument"></a>

DeleteDocument soft-deletes a document.

#### Method Signature

```protobuf
// Unary RPC
rpc DeleteDocument(DeleteDocumentRequest) returns (DeleteDocumentResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.DeleteDocument` |
| **Input Type** | [`DeleteDocumentRequest`](#deletedocumentrequest) |
| **Output Type** | [`DeleteDocumentResponse`](#deletedocumentresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: DeleteDocument
    Note right of Service: DeleteDocumentRequest
    Service-->>-Client: Response
    Note left of Client: DeleteDocumentResponse
```

---

### ListDocuments

<a name="listdocuments"></a>

ListDocuments lists documents with pagination and filtering.

#### Method Signature

```protobuf
// Unary RPC
rpc ListDocuments(ListDocumentsRequest) returns (ListDocumentsResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListDocuments` |
| **Input Type** | [`ListDocumentsRequest`](#listdocumentsrequest) |
| **Output Type** | [`ListDocumentsResponse`](#listdocumentsresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: ListDocuments
    Note right of Service: ListDocumentsRequest
    Service-->>-Client: Response
    Note left of Client: ListDocumentsResponse
```

---

### AddCollaborator

<a name="addcollaborator"></a>

AddCollaborator adds a collaborator to a document.

#### Method Signature

```protobuf
// Unary RPC
rpc AddCollaborator(AddCollaboratorRequest) returns (AddCollaboratorResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.AddCollaborator` |
| **Input Type** | [`AddCollaboratorRequest`](#addcollaboratorrequest) |
| **Output Type** | [`AddCollaboratorResponse`](#addcollaboratorresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: AddCollaborator
    Note right of Service: AddCollaboratorRequest
    Service-->>-Client: Response
    Note left of Client: AddCollaboratorResponse
```

---

### RemoveCollaborator

<a name="removecollaborator"></a>

RemoveCollaborator removes a collaborator from a document.

#### Method Signature

```protobuf
// Unary RPC
rpc RemoveCollaborator(RemoveCollaboratorRequest) returns (RemoveCollaboratorResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RemoveCollaborator` |
| **Input Type** | [`RemoveCollaboratorRequest`](#removecollaboratorrequest) |
| **Output Type** | [`RemoveCollaboratorResponse`](#removecollaboratorresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: RemoveCollaborator
    Note right of Service: RemoveCollaboratorRequest
    Service-->>-Client: Response
    Note left of Client: RemoveCollaboratorResponse
```

---

### GrantPermission

<a name="grantpermission"></a>

GrantPermission grants permission on a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc GrantPermission(GrantPermissionRequest) returns (GrantPermissionResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GrantPermission` |
| **Input Type** | [`GrantPermissionRequest`](#grantpermissionrequest) |
| **Output Type** | [`GrantPermissionResponse`](#grantpermissionresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GrantPermission
    Note right of Service: GrantPermissionRequest
    Service-->>-Client: Response
    Note left of Client: GrantPermissionResponse
```

---

### RevokePermission

<a name="revokepermission"></a>

RevokePermission revokes permission on a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc RevokePermission(RevokePermissionRequest) returns (RevokePermissionResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RevokePermission` |
| **Input Type** | [`RevokePermissionRequest`](#revokepermissionrequest) |
| **Output Type** | [`RevokePermissionResponse`](#revokepermissionresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: RevokePermission
    Note right of Service: RevokePermissionRequest
    Service-->>-Client: Response
    Note left of Client: RevokePermissionResponse
```

---

### CheckPermission

<a name="checkpermission"></a>

CheckPermission checks if a user has permission on a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc CheckPermission(CheckPermissionRequest) returns (CheckPermissionResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CheckPermission` |
| **Input Type** | [`CheckPermissionRequest`](#checkpermissionrequest) |
| **Output Type** | [`CheckPermissionResponse`](#checkpermissionresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CheckPermission
    Note right of Service: CheckPermissionRequest
    Service-->>-Client: Response
    Note left of Client: CheckPermissionResponse
```

---

### ListPermissions

<a name="listpermissions"></a>

ListPermissions lists permissions for a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc ListPermissions(ListPermissionsRequest) returns (ListPermissionsResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListPermissions` |
| **Input Type** | [`ListPermissionsRequest`](#listpermissionsrequest) |
| **Output Type** | [`ListPermissionsResponse`](#listpermissionsresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: ListPermissions
    Note right of Service: ListPermissionsRequest
    Service-->>-Client: Response
    Note left of Client: ListPermissionsResponse
```

---

### LogActivity

<a name="logactivity"></a>

LogActivity logs an activity on a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc LogActivity(LogActivityRequest) returns (LogActivityResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.LogActivity` |
| **Input Type** | [`LogActivityRequest`](#logactivityrequest) |
| **Output Type** | [`LogActivityResponse`](#logactivityresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: LogActivity
    Note right of Service: LogActivityRequest
    Service-->>-Client: Response
    Note left of Client: LogActivityResponse
```

---

### GetActivityLog

<a name="getactivitylog"></a>

GetActivityLog retrieves activity log for a resource.

#### Method Signature

```protobuf
// Unary RPC
rpc GetActivityLog(GetActivityLogRequest) returns (GetActivityLogResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetActivityLog` |
| **Input Type** | [`GetActivityLogRequest`](#getactivitylogrequest) |
| **Output Type** | [`GetActivityLogResponse`](#getactivitylogresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetActivityLog
    Note right of Service: GetActivityLogRequest
    Service-->>-Client: Response
    Note left of Client: GetActivityLogResponse
```

---

### GetQuota

<a name="getquota"></a>

GetQuota retrieves resource quota information.

#### Method Signature

```protobuf
// Unary RPC
rpc GetQuota(GetQuotaRequest) returns (GetQuotaResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetQuota` |
| **Input Type** | [`GetQuotaRequest`](#getquotarequest) |
| **Output Type** | [`GetQuotaResponse`](#getquotaresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetQuota
    Note right of Service: GetQuotaRequest
    Service-->>-Client: Response
    Note left of Client: GetQuotaResponse
```

---

### UpdateQuota

<a name="updatequota"></a>

UpdateQuota updates resource quota limits.

#### Method Signature

```protobuf
// Unary RPC
rpc UpdateQuota(UpdateQuotaRequest) returns (UpdateQuotaResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateQuota` |
| **Input Type** | [`UpdateQuotaRequest`](#updatequotarequest) |
| **Output Type** | [`UpdateQuotaResponse`](#updatequotaresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: UpdateQuota
    Note right of Service: UpdateQuotaRequest
    Service-->>-Client: Response
    Note left of Client: UpdateQuotaResponse
```

---

### StreamActivityFeed

<a name="streamactivityfeed"></a>

StreamActivityFeed streams real-time activity feed. Uses server-side streaming.

#### Method Signature

```protobuf
// Server streaming RPC
rpc StreamActivityFeed(StreamActivityFeedRequest) returns (stream ResourceActivity);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.StreamActivityFeed` |
| **Input Type** | [`StreamActivityFeedRequest`](#streamactivityfeedrequest) |
| **Output Type** | [`ResourceActivity`](#resourceactivity) |
| **Streaming Type** | Server Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Server Streaming
    Client->>+Service: StreamActivityFeed
    Client->>Service: StreamActivityFeedRequest
    loop Stream Messages
        Service-->>Client: ResourceActivity
    end
    Service-->>-Client: End Stream
```

---

### BatchGrantPermissions

<a name="batchgrantpermissions"></a>

BatchGrantPermissions grants multiple permissions. Uses client-side streaming.

#### Method Signature

```protobuf
// Client streaming RPC
rpc BatchGrantPermissions(stream GrantPermissionRequest) returns (BatchGrantPermissionsResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.BatchGrantPermissions` |
| **Input Type** | [`GrantPermissionRequest`](#grantpermissionrequest) |
| **Output Type** | [`BatchGrantPermissionsResponse`](#batchgrantpermissionsresponse) |
| **Streaming Type** | Client Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Client Streaming
    Client->>+Service: BatchGrantPermissions (stream)
    loop Stream Messages
        Client->>Service: GrantPermissionRequest
    end
    Service-->>-Client: BatchGrantPermissionsResponse
```

---

### CollaborateOnDocument

<a name="collaborateondocument"></a>

CollaborateOnDocument enables real-time document collaboration. Uses bidirectional streaming.

#### Method Signature

```protobuf
// Bidirectional streaming RPC
rpc CollaborateOnDocument(stream DocumentCollaboration) returns (stream DocumentCollaboration);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CollaborateOnDocument` |
| **Input Type** | [`DocumentCollaboration`](#documentcollaboration) |
| **Output Type** | [`DocumentCollaboration`](#documentcollaboration) |
| **Streaming Type** | Bidirectional Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Bidirectional Streaming
    Client->>+Service: CollaborateOnDocument (stream)
    loop Stream Messages
        Client->>Service: DocumentCollaboration
        Service-->>Client: DocumentCollaboration
    end
    Service-->>-Client: End Stream
```

---

## 📦 Messages

<a name="messages"></a>

This service defines **36 message types**:

### AddCollaboratorRequest

<a name="addcollaboratorrequest"></a>

AddCollaboratorRequest adds a collaborator to a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.AddCollaboratorRequest` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 2 | `user_id` | string | optional | User ID to add. (Must be a non-empty identifier) |
| 3 | `access_level` | [`AccessLevel`](#accesslevel) | optional | Access level. |

#### Proto Definition

```protobuf
message AddCollaboratorRequest {
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 1;
  // User ID to add. (Must be a non-empty identifier)
  optional string user_id = 2;
  // Access level.
  optional AccessLevel access_level = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AddCollaboratorRequest {
        +string document_id
        +string user_id
        +AccessLevel access_level
    }
    AddCollaboratorRequest --> AccessLevel
```

---

### GetQuotaResponse

<a name="getquotaresponse"></a>

GetQuotaResponse returns quota information.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetQuotaResponse` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `quota` | [`ResourceQuota`](#resourcequota) | optional | Resource quota. |
| 2 | `usage_percentage` | double | optional | Usage percentage Range: 0-100. |
| 3 | `over_quota` | bool | optional | Is over quota. |

#### Proto Definition

```protobuf
message GetQuotaResponse {
  // Resource quota.
  optional ResourceQuota quota = 1;
  // Usage percentage Range: 0-100.
  optional double usage_percentage = 2;
  // Is over quota.
  optional bool over_quota = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetQuotaResponse {
        +ResourceQuota quota
        +double usage_percentage
        +bool over_quota
    }
    GetQuotaResponse --> ResourceQuota
```

---

### BatchGrantPermissionsResponse

<a name="batchgrantpermissionsresponse"></a>

BatchGrantPermissionsResponse returns batch grant results.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.BatchGrantPermissionsResponse` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `permissions` | [`ResourcePermission`](#resourcepermission) | repeated | Granted permissions. |
| 2 | `granted_count` | int32 | optional | Number of permissions granted Must be >= 0. |
| 3 | `failed_count` | int32 | optional | Number of permissions failed Must be >= 0. |
| 4 | `errors` | [`Error`](#error) | repeated | Errors encountered. |

#### Proto Definition

```protobuf
message BatchGrantPermissionsResponse {
  // Granted permissions.
  repeated ResourcePermission permissions = 1;
  // Number of permissions granted Must be >= 0.
  optional int32 granted_count = 2;
  // Number of permissions failed Must be >= 0.
  optional int32 failed_count = 3;
  // Errors encountered.
  repeated Error errors = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class BatchGrantPermissionsResponse {
        +ResourcePermission[] permissions
        +int32 granted_count
        +int32 failed_count
        +Error[] errors
    }
    BatchGrantPermissionsResponse "1" --> "*" ResourcePermission
    BatchGrantPermissionsResponse "1" --> "*" Error
```

---

### DocumentCollaboration

<a name="documentcollaboration"></a>

DocumentCollaboration for real-time collaboration.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.DocumentCollaboration` |
| **Field Count** | 9 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `message_id` | string | optional | Message ID for tracking. (Must be a non-empty identifier) |
| 2 | `message_type` | string | optional | Message type (edit, cursor, selection, comment). |
| 3 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |
| 4 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 5 | `user_id` | string | optional | User ID. (Must be a non-empty identifier) |
| 6 | `content_change` | [`ContentChange`](#contentchange) | optional | Content change (for edit messages). |
| 7 | `cursor_position` | [`CursorPosition`](#cursorposition) | optional | Cursor position (for cursor messages). |
| 8 | `comment` | string | optional | Comment (for comment messages). |
| 9 | `ack_message_id` | string | optional | Acknowledge message ID. (Must be a non-empty identifier) |

#### Proto Definition

```protobuf
message DocumentCollaboration {
  // Message ID for tracking. (Must be a non-empty identifier)
  optional string message_id = 1;
  // Message type (edit, cursor, selection, comment).
  optional string message_type = 2;
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 3;
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 4;
  // User ID. (Must be a non-empty identifier)
  optional string user_id = 5;
  // Content change (for edit messages).
  optional ContentChange content_change = 6;
  // Cursor position (for cursor messages).
  optional CursorPosition cursor_position = 7;
  // Comment (for comment messages).
  optional string comment = 8;
  // Acknowledge message ID. (Must be a non-empty identifier)
  optional string ack_message_id = 9;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DocumentCollaboration {
        +string message_id
        +string message_type
        +Timestamp timestamp
        +string document_id
        +string user_id
        +ContentChange content_change
        +CursorPosition cursor_position
        +string comment
        +string ack_message_id
    }
    DocumentCollaboration --> Timestamp
    DocumentCollaboration --> ContentChange
    DocumentCollaboration --> CursorPosition
```

---

### CreateDocumentRequest

<a name="createdocumentrequest"></a>

CreateDocumentRequest creates a new document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CreateDocumentRequest` |
| **Field Count** | 7 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `project_id` | string | optional | Project ID (optional). (Must be a non-empty identifier) |
| 2 | `title` | string | optional | Document title. |
| 3 | `content` | string | optional | Document content. |
| 4 | `format` | string | optional | Document format (markdown, html, text). |
| 5 | `author_id` | string | optional | Author user ID. (Must be a non-empty identifier) |
| 6 | `collaborators` | [`DocumentCollaborator`](#documentcollaborator) | repeated | Initial collaborators. |
| 7 | `tags` | string | repeated | Tags. |

#### Proto Definition

```protobuf
message CreateDocumentRequest {
  // Project ID (optional). (Must be a non-empty identifier)
  optional string project_id = 1;
  // Document title.
  optional string title = 2;
  // Document content.
  optional string content = 3;
  // Document format (markdown, html, text).
  optional string format = 4;
  // Author user ID. (Must be a non-empty identifier)
  optional string author_id = 5;
  // Initial collaborators.
  repeated DocumentCollaborator collaborators = 6;
  // Tags.
  repeated string tags = 7;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateDocumentRequest {
        +string project_id
        +string title
        +string content
        +string format
        +string author_id
        +DocumentCollaborator[] collaborators
        +string[] tags
    }
    CreateDocumentRequest "1" --> "*" DocumentCollaborator
```

---

### UpdateDocumentResponse

<a name="updatedocumentresponse"></a>

UpdateDocumentResponse returns the updated document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateDocumentResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document` | [`Document`](#document) | optional | Updated document. |
| 2 | `new_version` | int32 | optional | New version number. |

#### Proto Definition

```protobuf
message UpdateDocumentResponse {
  // Updated document.
  optional Document document = 1;
  // New version number.
  optional int32 new_version = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateDocumentResponse {
        +Document document
        +int32 new_version
    }
    UpdateDocumentResponse --> Document
```

---

### RemoveCollaboratorRequest

<a name="removecollaboratorrequest"></a>

RemoveCollaboratorRequest removes a collaborator from a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RemoveCollaboratorRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 2 | `user_id` | string | optional | User ID to remove. (Must be a non-empty identifier) |

#### Proto Definition

```protobuf
message RemoveCollaboratorRequest {
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 1;
  // User ID to remove. (Must be a non-empty identifier)
  optional string user_id = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RemoveCollaboratorRequest {
        +string document_id
        +string user_id
    }
```

---

### ListPermissionsRequest

<a name="listpermissionsrequest"></a>

ListPermissionsRequest lists permissions.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListPermissionsRequest` |
| **Field Count** | 5 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_id` | string | optional | Filter by resource ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Filter by resource type. |
| 3 | `principal_id` | string | optional | Filter by principal ID. (Must be a non-empty identifier) |
| 4 | `include_expired` | bool | optional | Include expired permissions. |
| 5 | `pagination` | [`PaginationRequest`](#paginationrequest) | optional | Pagination. |

#### Proto Definition

```protobuf
message ListPermissionsRequest {
  // Filter by resource ID. (Must be a non-empty identifier)
  optional string resource_id = 1;
  // Filter by resource type.
  optional ResourceType resource_type = 2;
  // Filter by principal ID. (Must be a non-empty identifier)
  optional string principal_id = 3;
  // Include expired permissions.
  optional bool include_expired = 4;
  // Pagination.
  optional PaginationRequest pagination = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListPermissionsRequest {
        +string resource_id
        +ResourceType resource_type
        +string principal_id
        +bool include_expired
        +PaginationRequest pagination
    }
    ListPermissionsRequest --> ResourceType
    ListPermissionsRequest --> PaginationRequest
```

---

### ListPermissionsResponse

<a name="listpermissionsresponse"></a>

ListPermissionsResponse returns matching permissions.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListPermissionsResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `permissions` | [`ResourcePermission`](#resourcepermission) | repeated | Matching permissions. |
| 2 | `pagination` | [`PaginationResponse`](#paginationresponse) | optional | Pagination metadata. |

#### Proto Definition

```protobuf
message ListPermissionsResponse {
  // Matching permissions.
  repeated ResourcePermission permissions = 1;
  // Pagination metadata.
  optional PaginationResponse pagination = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListPermissionsResponse {
        +ResourcePermission[] permissions
        +PaginationResponse pagination
    }
    ListPermissionsResponse "1" --> "*" ResourcePermission
    ListPermissionsResponse --> PaginationResponse
```

---

### GetDocumentResponse

<a name="getdocumentresponse"></a>

GetDocumentResponse returns the requested document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetDocumentResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document` | [`Document`](#document) | optional | Retrieved document. |

#### Proto Definition

```protobuf
message GetDocumentResponse {
  // Retrieved document.
  optional Document document = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetDocumentResponse {
        +Document document
    }
    GetDocumentResponse --> Document
```

---

### DeleteDocumentRequest

<a name="deletedocumentrequest"></a>

DeleteDocumentRequest deletes a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.DeleteDocumentRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 2 | `hard_delete` | bool | optional | Hard delete (permanent). |

#### Proto Definition

```protobuf
message DeleteDocumentRequest {
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 1;
  // Hard delete (permanent).
  optional bool hard_delete = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeleteDocumentRequest {
        +string document_id
        +bool hard_delete
    }
```

---

### GrantPermissionRequest

<a name="grantpermissionrequest"></a>

GrantPermissionRequest grants permission on a resource.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GrantPermissionRequest` |
| **Field Count** | 7 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_id` | string | optional | Resource ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |
| 3 | `principal_id` | string | optional | Principal ID (user or group). (Must be a non-empty identifier) |
| 4 | `principal_type` | string | optional | Principal type (user or group). |
| 5 | `access_level` | [`AccessLevel`](#accesslevel) | optional | Access level. |
| 6 | `expires_at` | [`Timestamp`](#timestamp) | optional | Permission expiry. (RFC 3339 timestamp format) |
| 7 | `granted_by` | string | optional | Granted by user ID. |

#### Proto Definition

```protobuf
message GrantPermissionRequest {
  // Resource ID. (Must be a non-empty identifier)
  optional string resource_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
  // Principal ID (user or group). (Must be a non-empty identifier)
  optional string principal_id = 3;
  // Principal type (user or group).
  optional string principal_type = 4;
  // Access level.
  optional AccessLevel access_level = 5;
  // Permission expiry. (RFC 3339 timestamp format)
  optional Timestamp expires_at = 6;
  // Granted by user ID.
  optional string granted_by = 7;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GrantPermissionRequest {
        +string resource_id
        +ResourceType resource_type
        +string principal_id
        +string principal_type
        +AccessLevel access_level
        +Timestamp expires_at
        +string granted_by
    }
    GrantPermissionRequest --> ResourceType
    GrantPermissionRequest --> AccessLevel
    GrantPermissionRequest --> Timestamp
```

---

### LogActivityRequest

<a name="logactivityrequest"></a>

LogActivityRequest logs an activity.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.LogActivityRequest` |
| **Field Count** | 7 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_id` | string | optional | Resource ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |
| 3 | `activity_type` | string | optional | Activity type. |
| 4 | `user_id` | string | optional | User ID. (Must be a non-empty identifier) |
| 5 | `description` | string | optional | Description. |
| 6 | `details` | string | optional | Details (JSON). |
| 7 | `ip_address` | string | optional | IP address. |

#### Proto Definition

```protobuf
message LogActivityRequest {
  // Resource ID. (Must be a non-empty identifier)
  optional string resource_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
  // Activity type.
  optional string activity_type = 3;
  // User ID. (Must be a non-empty identifier)
  optional string user_id = 4;
  // Description.
  optional string description = 5;
  // Details (JSON).
  optional string details = 6;
  // IP address.
  optional string ip_address = 7;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class LogActivityRequest {
        +string resource_id
        +ResourceType resource_type
        +string activity_type
        +string user_id
        +string description
        +string details
        +string ip_address
    }
    LogActivityRequest --> ResourceType
```

---

### CursorPosition

<a name="cursorposition"></a>

CursorPosition represents a user's cursor position.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CursorPosition` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `line` | int32 | optional | Line number. |
| 2 | `column` | int32 | optional | Column number. |
| 3 | `selection_start` | [`Position`](#position) | optional | Selection start (if any). |
| 4 | `selection_end` | [`Position`](#position) | optional | Selection end (if any). |

#### Proto Definition

```protobuf
message CursorPosition {
  // Line number.
  optional int32 line = 1;
  // Column number.
  optional int32 column = 2;
  // Selection start (if any).
  optional Position selection_start = 3;
  // Selection end (if any).
  optional Position selection_end = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CursorPosition {
        +int32 line
        +int32 column
        +Position selection_start
        +Position selection_end
    }
    CursorPosition --> Position
    CursorPosition --> Position
```

---

### GetDocumentRequest

<a name="getdocumentrequest"></a>

GetDocumentRequest retrieves a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetDocumentRequest` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 2 | `include_deleted` | bool | optional | Include deleted documents. |
| 3 | `include_content` | bool | optional | Include content. |

#### Proto Definition

```protobuf
message GetDocumentRequest {
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 1;
  // Include deleted documents.
  optional bool include_deleted = 2;
  // Include content.
  optional bool include_content = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetDocumentRequest {
        +string document_id
        +bool include_deleted
        +bool include_content
    }
```

---

### ListDocumentsRequest

<a name="listdocumentsrequest"></a>

ListDocumentsRequest lists documents.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListDocumentsRequest` |
| **Field Count** | 11 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `pagination` | [`PaginationRequest`](#paginationrequest) | optional | Pagination. |
| 2 | `project_id` | string | optional | Filter by project ID. (Must be a non-empty identifier) |
| 3 | `author_id` | string | optional | Filter by author ID. (Must be a non-empty identifier) |
| 4 | `state` | [`ResourceState`](#resourcestate) | optional | Filter by state. |
| 5 | `format` | string | optional | Filter by format. |
| 6 | `tags` | string | repeated | Filter by tags. |
| 7 | `search_query` | string | optional | Search query. |
| 8 | `created_after` | [`Timestamp`](#timestamp) | optional | Created after date. |
| 9 | `created_before` | [`Timestamp`](#timestamp) | optional | Created before date. |
| 10 | `sort_by` | string | optional | Sort by field. |
| 11 | `sort_desc` | bool | optional | Sort descending. |

#### Proto Definition

```protobuf
message ListDocumentsRequest {
  // Pagination.
  optional PaginationRequest pagination = 1;
  // Filter by project ID. (Must be a non-empty identifier)
  optional string project_id = 2;
  // Filter by author ID. (Must be a non-empty identifier)
  optional string author_id = 3;
  // Filter by state.
  optional ResourceState state = 4;
  // Filter by format.
  optional string format = 5;
  // Filter by tags.
  repeated string tags = 6;
  // Search query.
  optional string search_query = 7;
  // Created after date.
  optional Timestamp created_after = 8;
  // Created before date.
  optional Timestamp created_before = 9;
  // Sort by field.
  optional string sort_by = 10;
  // Sort descending.
  optional bool sort_desc = 11;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListDocumentsRequest {
        +PaginationRequest pagination
        +string project_id
        +string author_id
        +ResourceState state
        +string format
        +string[] tags
        +string search_query
        +Timestamp created_after
        +Timestamp created_before
        +string sort_by
        +bool sort_desc
    }
    ListDocumentsRequest --> PaginationRequest
    ListDocumentsRequest --> ResourceState
    ListDocumentsRequest --> Timestamp
    ListDocumentsRequest --> Timestamp
```

---

### ListDocumentsResponse

<a name="listdocumentsresponse"></a>

ListDocumentsResponse returns matching documents.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ListDocumentsResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `documents` | [`Document`](#document) | repeated | Matching documents. |
| 2 | `pagination` | [`PaginationResponse`](#paginationresponse) | optional | Pagination metadata. |

#### Proto Definition

```protobuf
message ListDocumentsResponse {
  // Matching documents.
  repeated Document documents = 1;
  // Pagination metadata.
  optional PaginationResponse pagination = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ListDocumentsResponse {
        +Document[] documents
        +PaginationResponse pagination
    }
    ListDocumentsResponse "1" --> "*" Document
    ListDocumentsResponse --> PaginationResponse
```

---

### LogActivityResponse

<a name="logactivityresponse"></a>

LogActivityResponse returns the logged activity.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.LogActivityResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `activity` | [`ResourceActivity`](#resourceactivity) | optional | Logged activity. |

#### Proto Definition

```protobuf
message LogActivityResponse {
  // Logged activity.
  optional ResourceActivity activity = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class LogActivityResponse {
        +ResourceActivity activity
    }
    LogActivityResponse --> ResourceActivity
```

---

### CheckPermissionRequest

<a name="checkpermissionrequest"></a>

CheckPermissionRequest checks permission.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CheckPermissionRequest` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_id` | string | optional | Resource ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |
| 3 | `user_id` | string | optional | User ID to check. (Must be a non-empty identifier) |
| 4 | `required_access_level` | [`AccessLevel`](#accesslevel) | optional | Required access level. |

#### Proto Definition

```protobuf
message CheckPermissionRequest {
  // Resource ID. (Must be a non-empty identifier)
  optional string resource_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
  // User ID to check. (Must be a non-empty identifier)
  optional string user_id = 3;
  // Required access level.
  optional AccessLevel required_access_level = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CheckPermissionRequest {
        +string resource_id
        +ResourceType resource_type
        +string user_id
        +AccessLevel required_access_level
    }
    CheckPermissionRequest --> ResourceType
    CheckPermissionRequest --> AccessLevel
```

---

### CheckPermissionResponse

<a name="checkpermissionresponse"></a>

CheckPermissionResponse returns permission check result.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CheckPermissionResponse` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `has_permission` | bool | optional | Has permission. |
| 2 | `access_level` | [`AccessLevel`](#accesslevel) | optional | Actual access level. |
| 3 | `permission` | [`ResourcePermission`](#resourcepermission) | optional | Permission details. |

#### Proto Definition

```protobuf
message CheckPermissionResponse {
  // Has permission.
  optional bool has_permission = 1;
  // Actual access level.
  optional AccessLevel access_level = 2;
  // Permission details.
  optional ResourcePermission permission = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CheckPermissionResponse {
        +bool has_permission
        +AccessLevel access_level
        +ResourcePermission permission
    }
    CheckPermissionResponse --> AccessLevel
    CheckPermissionResponse --> ResourcePermission
```

---

### CreateDocumentResponse

<a name="createdocumentresponse"></a>

CreateDocumentResponse returns the created document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.CreateDocumentResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document` | [`Document`](#document) | optional | Created document. |

#### Proto Definition

```protobuf
message CreateDocumentResponse {
  // Created document.
  optional Document document = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateDocumentResponse {
        +Document document
    }
    CreateDocumentResponse --> Document
```

---

### GrantPermissionResponse

<a name="grantpermissionresponse"></a>

GrantPermissionResponse returns the granted permission.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GrantPermissionResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `permission` | [`ResourcePermission`](#resourcepermission) | optional | Granted permission. |

#### Proto Definition

```protobuf
message GrantPermissionResponse {
  // Granted permission.
  optional ResourcePermission permission = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GrantPermissionResponse {
        +ResourcePermission permission
    }
    GrantPermissionResponse --> ResourcePermission
```

---

### StreamActivityFeedRequest

<a name="streamactivityfeedrequest"></a>

StreamActivityFeedRequest requests activity feed stream.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.StreamActivityFeedRequest` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Filter by resource type. |
| 2 | `user_id` | string | optional | Filter by user ID. (Must be a non-empty identifier) |
| 3 | `activity_types` | string | repeated | Filter by activity types. |
| 4 | `start_from` | [`Timestamp`](#timestamp) | optional | Start from timestamp. |

#### Proto Definition

```protobuf
message StreamActivityFeedRequest {
  // Filter by resource type.
  optional ResourceType resource_type = 1;
  // Filter by user ID. (Must be a non-empty identifier)
  optional string user_id = 2;
  // Filter by activity types.
  repeated string activity_types = 3;
  // Start from timestamp.
  optional Timestamp start_from = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class StreamActivityFeedRequest {
        +ResourceType resource_type
        +string user_id
        +string[] activity_types
        +Timestamp start_from
    }
    StreamActivityFeedRequest --> ResourceType
    StreamActivityFeedRequest --> Timestamp
```

---

### GetQuotaRequest

<a name="getquotarequest"></a>

GetQuotaRequest retrieves quota information.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetQuotaRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `owner_id` | string | optional | Owner ID (user or organization). (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |

#### Proto Definition

```protobuf
message GetQuotaRequest {
  // Owner ID (user or organization). (Must be a non-empty identifier)
  optional string owner_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetQuotaRequest {
        +string owner_id
        +ResourceType resource_type
    }
    GetQuotaRequest --> ResourceType
```

---

### UpdateDocumentRequest

<a name="updatedocumentrequest"></a>

UpdateDocumentRequest updates a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateDocumentRequest` |
| **Field Count** | 6 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `document_id` | string | optional | Document ID. (Must be a non-empty identifier) |
| 2 | `title` | string | optional | Updated title. |
| 3 | `content` | string | optional | Updated content. |
| 4 | `format` | string | optional | Updated format. |
| 5 | `version` | int64 | optional | Version for optimistic locking. |
| 6 | `update_description` | string | optional | Update description. |

#### Proto Definition

```protobuf
message UpdateDocumentRequest {
  // Document ID. (Must be a non-empty identifier)
  optional string document_id = 1;
  // Updated title.
  optional string title = 2;
  // Updated content.
  optional string content = 3;
  // Updated format.
  optional string format = 4;
  // Version for optimistic locking.
  optional int64 version = 5;
  // Update description.
  optional string update_description = 6;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateDocumentRequest {
        +string document_id
        +string title
        +string content
        +string format
        +int64 version
        +string update_description
    }
```

---

### DeleteDocumentResponse

<a name="deletedocumentresponse"></a>

DeleteDocumentResponse confirms deletion.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.DeleteDocumentResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `deleted_at` | [`Timestamp`](#timestamp) | optional | Deletion timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message DeleteDocumentResponse {
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
    class DeleteDocumentResponse {
        +bool success
        +Timestamp deleted_at
    }
    DeleteDocumentResponse --> Timestamp
```

---

### AddCollaboratorResponse

<a name="addcollaboratorresponse"></a>

AddCollaboratorResponse returns the added collaborator.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.AddCollaboratorResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `collaborator` | [`DocumentCollaborator`](#documentcollaborator) | optional | Added collaborator. |
| 2 | `document` | [`Document`](#document) | optional | Updated document. |

#### Proto Definition

```protobuf
message AddCollaboratorResponse {
  // Added collaborator.
  optional DocumentCollaborator collaborator = 1;
  // Updated document.
  optional Document document = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AddCollaboratorResponse {
        +DocumentCollaborator collaborator
        +Document document
    }
    AddCollaboratorResponse --> DocumentCollaborator
    AddCollaboratorResponse --> Document
```

---

### RevokePermissionResponse

<a name="revokepermissionresponse"></a>

RevokePermissionResponse confirms revocation.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RevokePermissionResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `revoked_at` | [`Timestamp`](#timestamp) | optional | Revocation timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message RevokePermissionResponse {
  // Success status.
  optional bool success = 1;
  // Revocation timestamp. (RFC 3339 timestamp format)
  optional Timestamp revoked_at = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RevokePermissionResponse {
        +bool success
        +Timestamp revoked_at
    }
    RevokePermissionResponse --> Timestamp
```

---

### GetActivityLogResponse

<a name="getactivitylogresponse"></a>

GetActivityLogResponse returns activity log.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetActivityLogResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `activities` | [`ResourceActivity`](#resourceactivity) | repeated | Activity entries. |
| 2 | `pagination` | [`PaginationResponse`](#paginationresponse) | optional | Pagination metadata. |

#### Proto Definition

```protobuf
message GetActivityLogResponse {
  // Activity entries.
  repeated ResourceActivity activities = 1;
  // Pagination metadata.
  optional PaginationResponse pagination = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetActivityLogResponse {
        +ResourceActivity[] activities
        +PaginationResponse pagination
    }
    GetActivityLogResponse "1" --> "*" ResourceActivity
    GetActivityLogResponse --> PaginationResponse
```

---

### UpdateQuotaRequest

<a name="updatequotarequest"></a>

UpdateQuotaRequest updates quota limits.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateQuotaRequest` |
| **Field Count** | 5 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `owner_id` | string | optional | Owner ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |
| 3 | `max_count` | int64 | optional | New maximum count Must be >= 0. |
| 4 | `max_storage_bytes` | int64 | optional | New maximum storage bytes. |
| 5 | `updated_by` | string | optional | Updated by user ID. |

#### Proto Definition

```protobuf
message UpdateQuotaRequest {
  // Owner ID. (Must be a non-empty identifier)
  optional string owner_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
  // New maximum count Must be >= 0.
  optional int64 max_count = 3;
  // New maximum storage bytes.
  optional int64 max_storage_bytes = 4;
  // Updated by user ID.
  optional string updated_by = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateQuotaRequest {
        +string owner_id
        +ResourceType resource_type
        +int64 max_count
        +int64 max_storage_bytes
        +string updated_by
    }
    UpdateQuotaRequest --> ResourceType
```

---

### RemoveCollaboratorResponse

<a name="removecollaboratorresponse"></a>

RemoveCollaboratorResponse confirms removal.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RemoveCollaboratorResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `success` | bool | optional | Success status. |
| 2 | `document` | [`Document`](#document) | optional | Updated document. |

#### Proto Definition

```protobuf
message RemoveCollaboratorResponse {
  // Success status.
  optional bool success = 1;
  // Updated document.
  optional Document document = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RemoveCollaboratorResponse {
        +bool success
        +Document document
    }
    RemoveCollaboratorResponse --> Document
```

---

### RevokePermissionRequest

<a name="revokepermissionrequest"></a>

RevokePermissionRequest revokes a permission.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.RevokePermissionRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `permission_id` | string | optional | Permission ID. (Must be a non-empty identifier) |
| 2 | `revoked_by` | string | optional | Revoked by user ID. |

#### Proto Definition

```protobuf
message RevokePermissionRequest {
  // Permission ID. (Must be a non-empty identifier)
  optional string permission_id = 1;
  // Revoked by user ID.
  optional string revoked_by = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class RevokePermissionRequest {
        +string permission_id
        +string revoked_by
    }
```

---

### GetActivityLogRequest

<a name="getactivitylogrequest"></a>

GetActivityLogRequest retrieves activity log.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.GetActivityLogRequest` |
| **Field Count** | 7 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `resource_id` | string | optional | Resource ID. (Must be a non-empty identifier) |
| 2 | `resource_type` | [`ResourceType`](#resourcetype) | optional | Resource type. |
| 3 | `user_id` | string | optional | Filter by user ID. (Must be a non-empty identifier) |
| 4 | `activity_type` | string | optional | Filter by activity type. |
| 5 | `after` | [`Timestamp`](#timestamp) | optional | Activities after timestamp. |
| 6 | `before` | [`Timestamp`](#timestamp) | optional | Activities before timestamp. |
| 7 | `pagination` | [`PaginationRequest`](#paginationrequest) | optional | Pagination. |

#### Proto Definition

```protobuf
message GetActivityLogRequest {
  // Resource ID. (Must be a non-empty identifier)
  optional string resource_id = 1;
  // Resource type.
  optional ResourceType resource_type = 2;
  // Filter by user ID. (Must be a non-empty identifier)
  optional string user_id = 3;
  // Filter by activity type.
  optional string activity_type = 4;
  // Activities after timestamp.
  optional Timestamp after = 5;
  // Activities before timestamp.
  optional Timestamp before = 6;
  // Pagination.
  optional PaginationRequest pagination = 7;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetActivityLogRequest {
        +string resource_id
        +ResourceType resource_type
        +string user_id
        +string activity_type
        +Timestamp after
        +Timestamp before
        +PaginationRequest pagination
    }
    GetActivityLogRequest --> ResourceType
    GetActivityLogRequest --> Timestamp
    GetActivityLogRequest --> Timestamp
    GetActivityLogRequest --> PaginationRequest
```

---

### UpdateQuotaResponse

<a name="updatequotaresponse"></a>

UpdateQuotaResponse returns the updated quota.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.UpdateQuotaResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `quota` | [`ResourceQuota`](#resourcequota) | optional | Updated quota. |

#### Proto Definition

```protobuf
message UpdateQuotaResponse {
  // Updated quota.
  optional ResourceQuota quota = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UpdateQuotaResponse {
        +ResourceQuota quota
    }
    UpdateQuotaResponse --> ResourceQuota
```

---

### ContentChange

<a name="contentchange"></a>

ContentChange represents a change to document content.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.ContentChange` |
| **Field Count** | 5 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `operation` | string | optional | Change operation (insert, delete, replace). |
| 2 | `start_position` | int32 | optional | Start position. |
| 3 | `end_position` | int32 | optional | End position. |
| 4 | `content` | string | optional | New content. |
| 5 | `version` | int32 | optional | Version before change. |

#### Proto Definition

```protobuf
message ContentChange {
  // Change operation (insert, delete, replace).
  optional string operation = 1;
  // Start position.
  optional int32 start_position = 2;
  // End position.
  optional int32 end_position = 3;
  // New content.
  optional string content = 4;
  // Version before change.
  optional int32 version = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class ContentChange {
        +string operation
        +int32 start_position
        +int32 end_position
        +string content
        +int32 version
    }
```

---

### Position

<a name="position"></a>

Position represents a position in a document.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `second.v1.Position` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `line` | int32 | optional | Line number. |
| 2 | `column` | int32 | optional | Column number. |

#### Proto Definition

```protobuf
message Position {
  // Line number.
  optional int32 line = 1;
  // Column number.
  optional int32 column = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Position {
        +int32 line
        +int32 column
    }
```

---

## 🗄️ Data Model (ERD)

<a name="erd"></a>

Entity-Relationship diagram showing the data model.

```mermaid
%{init: {'theme':'forest'}}%
erDiagram
    AddCollaboratorRequest {
        string document_id
        string user_id
        AccessLevel access_level
    }

    AddCollaboratorRequest ||--|| AccessLevel : has
    GetQuotaResponse {
        ResourceQuota quota
        double usage_percentage
        bool over_quota
    }

    GetQuotaResponse ||--|| ResourceQuota : has
    BatchGrantPermissionsResponse {
        ResourcePermission permissions
        int32 granted_count
        int32 failed_count
        Error errors
    }

    BatchGrantPermissionsResponse ||--o{ ResourcePermission : has
    BatchGrantPermissionsResponse ||--o{ Error : has
    DocumentCollaboration {
        string message_id
        string message_type
        Timestamp timestamp
        string document_id
        string user_id
        ContentChange content_change
        CursorPosition cursor_position
        string comment
        string ack_message_id
    }

    DocumentCollaboration ||--|| ContentChange : has
    DocumentCollaboration ||--|| CursorPosition : has
    CreateDocumentRequest {
        string project_id
        string title
        string content
        string format
        string author_id
        DocumentCollaborator collaborators
        string tags
    }

    CreateDocumentRequest ||--o{ DocumentCollaborator : has
    UpdateDocumentResponse {
        Document document
        int32 new_version
    }

    UpdateDocumentResponse ||--|| Document : has
    RemoveCollaboratorRequest {
        string document_id
        string user_id
    }

    ListPermissionsRequest {
        string resource_id
        ResourceType resource_type
        string principal_id
        bool include_expired
        PaginationRequest pagination
    }

    ListPermissionsRequest ||--|| ResourceType : has
    ListPermissionsRequest ||--|| PaginationRequest : has
    ListPermissionsResponse {
        ResourcePermission permissions
        PaginationResponse pagination
    }

    ListPermissionsResponse ||--o{ ResourcePermission : has
    ListPermissionsResponse ||--|| PaginationResponse : has
    GetDocumentResponse {
        Document document
    }

    GetDocumentResponse ||--|| Document : has
    DeleteDocumentRequest {
        string document_id
        bool hard_delete
    }

    GrantPermissionRequest {
        string resource_id
        ResourceType resource_type
        string principal_id
        string principal_type
        AccessLevel access_level
        Timestamp expires_at
        string granted_by
    }

    GrantPermissionRequest ||--|| ResourceType : has
    GrantPermissionRequest ||--|| AccessLevel : has
    LogActivityRequest {
        string resource_id
        ResourceType resource_type
        string activity_type
        string user_id
        string description
        string details
        string ip_address
    }

    LogActivityRequest ||--|| ResourceType : has
    CursorPosition {
        int32 line
        int32 column
        Position selection_start
        Position selection_end
    }

    CursorPosition ||--|| Position : has
    CursorPosition ||--|| Position : has
    GetDocumentRequest {
        string document_id
        bool include_deleted
        bool include_content
    }

    ListDocumentsRequest {
        PaginationRequest pagination
        string project_id
        string author_id
        ResourceState state
        string format
        string tags
        string search_query
        Timestamp created_after
        Timestamp created_before
        string sort_by
        bool sort_desc
    }

    ListDocumentsRequest ||--|| PaginationRequest : has
    ListDocumentsRequest ||--|| ResourceState : has
    ListDocumentsResponse {
        Document documents
        PaginationResponse pagination
    }

    ListDocumentsResponse ||--o{ Document : has
    ListDocumentsResponse ||--|| PaginationResponse : has
    LogActivityResponse {
        ResourceActivity activity
    }

    LogActivityResponse ||--|| ResourceActivity : has
    CheckPermissionRequest {
        string resource_id
        ResourceType resource_type
        string user_id
        AccessLevel required_access_level
    }

    CheckPermissionRequest ||--|| ResourceType : has
    CheckPermissionRequest ||--|| AccessLevel : has
    CheckPermissionResponse {
        bool has_permission
        AccessLevel access_level
        ResourcePermission permission
    }

    CheckPermissionResponse ||--|| AccessLevel : has
    CheckPermissionResponse ||--|| ResourcePermission : has
    CreateDocumentResponse {
        Document document
    }

    CreateDocumentResponse ||--|| Document : has
    GrantPermissionResponse {
        ResourcePermission permission
    }

    GrantPermissionResponse ||--|| ResourcePermission : has
    StreamActivityFeedRequest {
        ResourceType resource_type
        string user_id
        string activity_types
        Timestamp start_from
    }

    StreamActivityFeedRequest ||--|| ResourceType : has
    GetQuotaRequest {
        string owner_id
        ResourceType resource_type
    }

    GetQuotaRequest ||--|| ResourceType : has
    UpdateDocumentRequest {
        string document_id
        string title
        string content
        string format
        int64 version
        string update_description
    }

    DeleteDocumentResponse {
        bool success
        Timestamp deleted_at
    }

    AddCollaboratorResponse {
        DocumentCollaborator collaborator
        Document document
    }

    AddCollaboratorResponse ||--|| DocumentCollaborator : has
    AddCollaboratorResponse ||--|| Document : has
    RevokePermissionResponse {
        bool success
        Timestamp revoked_at
    }

    GetActivityLogResponse {
        ResourceActivity activities
        PaginationResponse pagination
    }

    GetActivityLogResponse ||--o{ ResourceActivity : has
    GetActivityLogResponse ||--|| PaginationResponse : has
    UpdateQuotaRequest {
        string owner_id
        ResourceType resource_type
        int64 max_count
        int64 max_storage_bytes
        string updated_by
    }

    UpdateQuotaRequest ||--|| ResourceType : has
    RemoveCollaboratorResponse {
        bool success
        Document document
    }

    RemoveCollaboratorResponse ||--|| Document : has
    RevokePermissionRequest {
        string permission_id
        string revoked_by
    }

    GetActivityLogRequest {
        string resource_id
        ResourceType resource_type
        string user_id
        string activity_type
        Timestamp after
        Timestamp before
        PaginationRequest pagination
    }

    GetActivityLogRequest ||--|| ResourceType : has
    GetActivityLogRequest ||--|| PaginationRequest : has
    UpdateQuotaResponse {
        ResourceQuota quota
    }

    UpdateQuotaResponse ||--|| ResourceQuota : has
    ContentChange {
        string operation
        int32 start_position
        int32 end_position
        string content
        int32 version
    }

    Position {
        int32 line
        int32 column
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

    pb "second.v1"
)

func main() {
    // Connect to the service
    conn, err := grpc.Dial("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewSecondServiceClient(conn)

    // Example RPC call
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &pb.CreateDocumentRequest{
        // Fill in request fields
    }

    resp, err := client.CreateDocument(ctx, req)
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
import { ProtoGrpcType } from './second/second';
import { SecondServiceClient } from './second.v1/SecondService';

// Load proto file
const packageDefinition = protoLoader.loadSync(
    'second/second.proto',
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
const client: SecondServiceClient = new proto.second.v1.SecondService(
    'localhost:50051',
    grpc.credentials.createInsecure()
);

// Example RPC call
const request = {
    // Fill in request fields
};

client.CreateDocument(request, (error: grpc.ServiceError | null, response?: any) => {
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
| Generated At | 2025-11-23 00:35:25 UTC |
| Generator Version | 7.0.0 |

📚 **Documentation** | 🔧 **ProtoDocs** | ✨ **Auto-Generated**

</div>
