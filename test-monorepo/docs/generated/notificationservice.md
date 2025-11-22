# NotificationService

NotificationService service

**Package:** `notifications.v1`

**Version:** 1.0

## Methods

### Create

Creates a new NotificationService resource

**Input:** `CreateNotificationServiceRequest`

**Output:** `CreateNotificationServiceResponse`

### Get

Retrieves a NotificationService resource by ID

**Input:** `GetNotificationServiceRequest`

**Output:** `GetNotificationServiceResponse`

### List

Lists NotificationService resources with pagination

**Input:** `ListNotificationServiceRequest`

**Output:** `ListNotificationServiceResponse`

**Streaming:** Server

## Messages

### CreateNotificationServiceRequest

Request message for creating a NotificationService

| Field | Type | Description |
|-------|------|-------------|
| name | `string` | Resource name |
| description | `string` | Resource description |

### CreateNotificationServiceResponse

Response message for creating a NotificationService

| Field | Type | Description |
|-------|------|-------------|
| id | `string` | Created resource ID |
| status | `string` | Creation status |

