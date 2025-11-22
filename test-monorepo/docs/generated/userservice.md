# UserService

UserService service

**Package:** `users.v1`

**Version:** 1.0

## Methods

### Create

Creates a new UserService resource

**Input:** `CreateUserServiceRequest`

**Output:** `CreateUserServiceResponse`

### Get

Retrieves a UserService resource by ID

**Input:** `GetUserServiceRequest`

**Output:** `GetUserServiceResponse`

### List

Lists UserService resources with pagination

**Input:** `ListUserServiceRequest`

**Output:** `ListUserServiceResponse`

**Streaming:** Server

## Messages

### CreateUserServiceRequest

Request message for creating a UserService

| Field | Type | Description |
|-------|------|-------------|
| name | `string` | Resource name |
| description | `string` | Resource description |

### CreateUserServiceResponse

Response message for creating a UserService

| Field | Type | Description |
|-------|------|-------------|
| id | `string` | Created resource ID |
| status | `string` | Creation status |

