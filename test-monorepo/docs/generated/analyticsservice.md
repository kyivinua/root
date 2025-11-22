# AnalyticsService

AnalyticsService service

**Package:** `analytics.v1`

**Version:** 1.0

## Methods

### Create

Creates a new AnalyticsService resource

**Input:** `CreateAnalyticsServiceRequest`

**Output:** `CreateAnalyticsServiceResponse`

### Get

Retrieves a AnalyticsService resource by ID

**Input:** `GetAnalyticsServiceRequest`

**Output:** `GetAnalyticsServiceResponse`

### List

Lists AnalyticsService resources with pagination

**Input:** `ListAnalyticsServiceRequest`

**Output:** `ListAnalyticsServiceResponse`

**Streaming:** Server

## Messages

### CreateAnalyticsServiceRequest

Request message for creating a AnalyticsService

| Field | Type | Description |
|-------|------|-------------|
| name | `string` | Resource name |
| description | `string` | Resource description |

### CreateAnalyticsServiceResponse

Response message for creating a AnalyticsService

| Field | Type | Description |
|-------|------|-------------|
| id | `string` | Created resource ID |
| status | `string` | Creation status |

