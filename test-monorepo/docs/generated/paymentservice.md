# PaymentService

PaymentService service

**Package:** `payments.v1`

**Version:** 1.0

## Methods

### Create

Creates a new PaymentService resource

**Input:** `CreatePaymentServiceRequest`

**Output:** `CreatePaymentServiceResponse`

### Get

Retrieves a PaymentService resource by ID

**Input:** `GetPaymentServiceRequest`

**Output:** `GetPaymentServiceResponse`

### List

Lists PaymentService resources with pagination

**Input:** `ListPaymentServiceRequest`

**Output:** `ListPaymentServiceResponse`

**Streaming:** Server

## Messages

### CreatePaymentServiceRequest

Request message for creating a PaymentService

| Field | Type | Description |
|-------|------|-------------|
| name | `string` | Resource name |
| description | `string` | Resource description |

### CreatePaymentServiceResponse

Response message for creating a PaymentService

| Field | Type | Description |
|-------|------|-------------|
| id | `string` | Created resource ID |
| status | `string` | Creation status |

