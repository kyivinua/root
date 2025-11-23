# 📚 AnalyticsService API Documentation

| **Attribute** | **Value** |
|---------------|----------|
| **Service Name** | `AnalyticsService` |
| **Package** | `analytics.v1` |
| **Version** | v1 |
| **Proto File** | `analytics/analytics.proto` |
| **Generated** | 2025-11-23T00:35:25Z |

AnalyticsService provides analytics and reporting.

---

## 📑 Table of Contents

- [Overview](#overview)
- [Architecture](#architecture)
- [gRPC Service Interactions](#service-interaction)
- [Message Type Diagrams](#class-diagram)
- [Methods](#methods)
  - [TrackEvent](#trackevent)
  - [BatchTrackEvents](#batchtrackevents)
  - [GetMetrics](#getmetrics)
  - [GetReport](#getreport)
  - [StreamMetrics](#streammetrics)
  - [QueryData](#querydata)
  - [CreateDashboard](#createdashboard)
  - [GetDashboard](#getdashboard)
- [Messages](#messages)
  - [DataPoint](#datapoint)
  - [GetReportResponse](#getreportresponse)
  - [QueryResponse](#queryresponse)
  - [GetDashboardResponse](#getdashboardresponse)
  - [DataTable](#datatable)
  - [Dimension](#dimension)
  - [Row](#row)
  - [Position](#position)
  - [CreateDashboardRequest](#createdashboardrequest)
  - [Metric](#metric)
  - [Dashboard](#dashboard)
  - [DeviceContext](#devicecontext)
  - [LocationContext](#locationcontext)
  - [UTMContext](#utmcontext)
  - [Widget](#widget)
  - [Layout](#layout)
  - [TrackEventResponse](#trackeventresponse)
  - [QueryRequest](#queryrequest)
  - [MetricSummary](#metricsummary)
  - [DimensionValue](#dimensionvalue)
  - [GetMetricsResponse](#getmetricsresponse)
  - [StreamMetricsRequest](#streammetricsrequest)
  - [MetricUpdate](#metricupdate)
  - [CreateDashboardResponse](#createdashboardresponse)
  - [TimeRange](#timerange)
  - [Resolution](#resolution)
  - [Size](#size)
  - [SessionContext](#sessioncontext)
  - [BatchTrackResponse](#batchtrackresponse)
  - [GetDashboardRequest](#getdashboardrequest)
  - [Event](#event)
  - [Report](#report)
  - [Column](#column)
  - [GetReportRequest](#getreportrequest)
  - [UserContext](#usercontext)
  - [GetMetricsRequest](#getmetricsrequest)
  - [MetricSeries](#metricseries)
  - [Cell](#cell)
  - [TrackEventRequest](#trackeventrequest)
  - [Chart](#chart)
- [Enumerations](#enumerations)
- [Error Codes](#error-codes)
- [Examples](#examples)

---

## 📖 Overview

<a name="overview"></a>

### Service Statistics

| Metric | Count |
|--------|-------|
| **RPC Methods** | 8 |
| **Message Types** | 40 |
| **Enumerations** | 10 |
| **Streaming RPCs** | 2 |

### Quick Start

This service provides the following capabilities:

- [`TrackEvent`](#trackevent): TrackEvent tracks a single analytics event.
- [`BatchTrackEvents`](#batchtrackevents) (client streaming): BatchTrackEvents tracks multiple events. Uses client-side streaming to send multiple requests
- [`GetMetrics`](#getmetrics): GetMetrics retrieves metrics.
- [`GetReport`](#getreport): GetReport generates a report.
- [`StreamMetrics`](#streammetrics) (server streaming): StreamMetrics streams real-time metrics.
- ... and 3 more methods

---

## 🏗️ Architecture

<a name="architecture"></a>

```mermaid
%{init: {'theme':'forest'}}%
graph TB
    classDef serviceClass fill:#e1f5ff,stroke:#01579b,stroke-width:3px
    classDef methodClass fill:#f3e5f5,stroke:#4a148c,stroke-width:2px
    classDef messageClass fill:#e8f5e9,stroke:#1b5e20,stroke-width:2px

    AnalyticsService[🔧 AnalyticsService]:::serviceClass

    TrackEvent[TrackEvent]:::methodClass
    AnalyticsService --> TrackEvent
    TrackEvent_in[📥 TrackEventRequest]:::messageClass
    TrackEvent_out[📤 TrackEventResponse]:::messageClass
    TrackEvent_in -.->|input| TrackEvent
    TrackEvent -.->|output| TrackEvent_out
    BatchTrackEvents[↑ BatchTrackEvents]:::methodClass
    AnalyticsService --> BatchTrackEvents
    BatchTrackEvents_in[📥 TrackEventRequest]:::messageClass
    BatchTrackEvents_out[📤 BatchTrackResponse]:::messageClass
    BatchTrackEvents_in -.->|input| BatchTrackEvents
    BatchTrackEvents -.->|output| BatchTrackEvents_out
    GetMetrics[GetMetrics]:::methodClass
    AnalyticsService --> GetMetrics
    GetMetrics_in[📥 GetMetricsRequest]:::messageClass
    GetMetrics_out[📤 GetMetricsResponse]:::messageClass
    GetMetrics_in -.->|input| GetMetrics
    GetMetrics -.->|output| GetMetrics_out
    GetReport[GetReport]:::methodClass
    AnalyticsService --> GetReport
    GetReport_in[📥 GetReportRequest]:::messageClass
    GetReport_out[📤 GetReportResponse]:::messageClass
    GetReport_in -.->|input| GetReport
    GetReport -.->|output| GetReport_out
    StreamMetrics[↓ StreamMetrics]:::methodClass
    AnalyticsService --> StreamMetrics
    StreamMetrics_in[📥 StreamMetricsRequest]:::messageClass
    StreamMetrics_out[📤 MetricUpdate]:::messageClass
    StreamMetrics_in -.->|input| StreamMetrics
    StreamMetrics -.->|output| StreamMetrics_out
    QueryData[QueryData]:::methodClass
    AnalyticsService --> QueryData
    QueryData_in[📥 QueryRequest]:::messageClass
    QueryData_out[📤 QueryResponse]:::messageClass
    QueryData_in -.->|input| QueryData
    QueryData -.->|output| QueryData_out
    CreateDashboard[CreateDashboard]:::methodClass
    AnalyticsService --> CreateDashboard
    CreateDashboard_in[📥 CreateDashboardRequest]:::messageClass
    CreateDashboard_out[📤 CreateDashboardResponse]:::messageClass
    CreateDashboard_in -.->|input| CreateDashboard
    CreateDashboard -.->|output| CreateDashboard_out
    GetDashboard[GetDashboard]:::methodClass
    AnalyticsService --> GetDashboard
    GetDashboard_in[📥 GetDashboardRequest]:::messageClass
    GetDashboard_out[📤 GetDashboardResponse]:::messageClass
    GetDashboard_in -.->|input| GetDashboard
    GetDashboard -.->|output| GetDashboard_out
```

---

## 🔄 gRPC Service Interactions

<a name="service-interaction"></a>

This diagram shows the interactions between the service methods and message types.

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant AnalyticsService
    Client->>+AnalyticsService: TrackEvent
    Note right of AnalyticsService: TrackEventRequest
    AnalyticsService->>-Client: TrackEventResponse
    Client->>+AnalyticsService: BatchTrackEvents (client stream)
    Note over Client,AnalyticsService: Stream of TrackEventRequest
    AnalyticsService->>-Client: BatchTrackResponse
    Client->>+AnalyticsService: GetMetrics
    Note right of AnalyticsService: GetMetricsRequest
    AnalyticsService->>-Client: GetMetricsResponse
    Client->>+AnalyticsService: GetReport
    Note right of AnalyticsService: GetReportRequest
    AnalyticsService->>-Client: GetReportResponse
    Client->>+AnalyticsService: StreamMetrics
    Note right of AnalyticsService: StreamMetricsRequest
    AnalyticsService->>-Client: Stream of MetricUpdate
    Client->>+AnalyticsService: QueryData
    Note right of AnalyticsService: QueryRequest
    AnalyticsService->>-Client: QueryResponse
    Client->>+AnalyticsService: CreateDashboard
    Note right of AnalyticsService: CreateDashboardRequest
    AnalyticsService->>-Client: CreateDashboardResponse
    Client->>+AnalyticsService: GetDashboard
    Note right of AnalyticsService: GetDashboardRequest
    AnalyticsService->>-Client: GetDashboardResponse
```

---

## 📦 Message Type Diagrams

<a name="class-diagram"></a>

UML class diagrams showing the structure of message types.

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class AnalyticsService {
        <<service>>
        +DataPoint()
        +GetReportResponse()
        +QueryResponse()
        +GetDashboardResponse()
        +DataTable()
        +Dimension()
        +Row()
        +Position()
        +CreateDashboardRequest()
        +Metric()
        +Dashboard()
        +DeviceContext()
        +LocationContext()
        +UTMContext()
        +Widget()
        +Layout()
        +TrackEventResponse()
        +QueryRequest()
        +MetricSummary()
        +DimensionValue()
        +GetMetricsResponse()
        +StreamMetricsRequest()
        +MetricUpdate()
        +CreateDashboardResponse()
        +TimeRange()
        +Resolution()
        +Size()
        +SessionContext()
        +BatchTrackResponse()
        +GetDashboardRequest()
        +Event()
        +Report()
        +Column()
        +GetReportRequest()
        +UserContext()
        +GetMetricsRequest()
        +MetricSeries()
        +Cell()
        +TrackEventRequest()
        +Chart()
    }

    class DataPoint {
        +Timestamp timestamp
        +double value
        +map<string, string> labels
    }

    class GetReportResponse {
        +Report report
    }

    GetReportResponse "1" --> "1" Report
    class QueryResponse {
        +DataTable result
        +Duration execution_time
    }

    QueryResponse "1" --> "1" DataTable
    class GetDashboardResponse {
        +Dashboard dashboard
        +map<string, WidgetData> widget_data
    }

    GetDashboardResponse "1" --> "1" Dashboard
    class DataTable {
        +string name
        +Column columns[]
        +Row rows[]
        +int64 total_count
    }

    DataTable "1" --> "*" Column
    DataTable "1" --> "*" Row
    class Dimension {
        +string name
        +DimensionValue values[]
        +int64 total_count
    }

    Dimension "1" --> "*" DimensionValue
    class Row {
        +Cell cells[]
    }

    Row "1" --> "*" Cell
    class Position {
        +int32 x
        +int32 y
    }

    class CreateDashboardRequest {
        +Dashboard dashboard
    }

    CreateDashboardRequest "1" --> "1" Dashboard
    class Metric {
        +string name
        +MetricType type
        +Timestamp timestamp
        +double value
        +map<string, string> tags
        +string unit
    }

    Metric "1" --> "1" MetricType
    class Dashboard {
        +string dashboard_id
        +string name
        +string description
        +string owner_id
        +Widget widgets[]
        +Layout layout
        +Duration refresh_interval
        +Timestamp created_at
        +Timestamp updated_at
    }

    Dashboard "1" --> "*" Widget
    Dashboard "1" --> "1" Layout
    class DeviceContext {
        +DeviceType device_type
        +string os
        +string os_version
        +string browser
        +string browser_version
        +string brand
        +string model
        +Resolution screen
        +string user_agent
    }

    DeviceContext "1" --> "1" DeviceType
    DeviceContext "1" --> "1" Resolution
    class LocationContext {
        +string ip
        +string country
        +string region
        +string city
        +string postal_code
        +double latitude
        +double longitude
        +string timezone
    }

    class UTMContext {
        +string source
        +string medium
        +string campaign
        +string term
        +string content
    }

    class Widget {
        +string widget_id
        +WidgetType type
        +string title
        +Position position
        +Size size
        +map<string, string> config
        +QueryRequest query
    }

    Widget "1" --> "1" WidgetType
    Widget "1" --> "1" Position
    Widget "1" --> "1" Size
    Widget "1" --> "1" QueryRequest
    class Layout {
        +int32 columns
        +int32 rows
        +int32 grid_size
    }

    class TrackEventResponse {
        +string event_id
        +bool success
    }

    class QueryRequest {
        +string query
        +map<string, string> parameters
        +int32 limit
        +int32 offset
    }

    class MetricSummary {
        +string name
        +double current_value
        +double previous_value
        +double change_percent
        +Trend trend
        +AggregationType aggregation
    }

    MetricSummary "1" --> "1" Trend
    MetricSummary "1" --> "1" AggregationType
    class DimensionValue {
        +string value
        +int64 count
        +double percentage
        +map<string, double> metrics
    }

    class GetMetricsResponse {
        +MetricSeries metrics[]
    }

    GetMetricsResponse "1" --> "*" MetricSeries
    class StreamMetricsRequest {
        +string metric_names[]
        +Duration interval
    }

    class MetricUpdate {
        +Metric metric
        +Timestamp timestamp
    }

    MetricUpdate "1" --> "1" Metric
    class CreateDashboardResponse {
        +Dashboard dashboard
    }

    CreateDashboardResponse "1" --> "1" Dashboard
    class TimeRange {
        +Timestamp start
        +Timestamp end
        +TimeGranularity granularity
    }

    TimeRange "1" --> "1" TimeGranularity
    class Resolution {
        +int32 width
        +int32 height
    }

    class Size {
        +int32 width
        +int32 height
    }

    class SessionContext {
        +string session_id
        +Timestamp started_at
        +Duration duration
        +int32 page_views
        +int32 event_count
    }

    class BatchTrackResponse {
        +int32 events_tracked
        +int32 failed_count
    }

    class GetDashboardRequest {
        +string dashboard_id
    }

    class Event {
        +string event_id
        +EventType event_type
        +string event_name
        +Timestamp timestamp
        +UserContext user
        +SessionContext session
        +DeviceContext device
        +LocationContext location
        +map<string, string> properties
        +double value
        +string currency
        +UTMContext utm
        +string referrer
        +map<string, string> dimensions
    }

    Event "1" --> "1" EventType
    Event "1" --> "1" UserContext
    Event "1" --> "1" SessionContext
    Event "1" --> "1" DeviceContext
    Event "1" --> "1" LocationContext
    Event "1" --> "1" UTMContext
    class Report {
        +string report_id
        +string name
        +ReportType type
        +TimeRange time_range
        +MetricSummary metrics[]
        +Dimension dimensions[]
        +Chart charts[]
        +DataTable tables[]
        +Timestamp generated_at
    }

    Report "1" --> "1" ReportType
    Report "1" --> "1" TimeRange
    Report "1" --> "*" MetricSummary
    Report "1" --> "*" Dimension
    Report "1" --> "*" Chart
    Report "1" --> "*" DataTable
    class Column {
        +string name
        +DataType type
        +string format
    }

    Column "1" --> "1" DataType
    class GetReportRequest {
        +ReportType type
        +TimeRange time_range
        +map<string, string> filters
        +string metrics[]
        +string dimensions[]
    }

    GetReportRequest "1" --> "1" ReportType
    GetReportRequest "1" --> "1" TimeRange
    class UserContext {
        +string user_id
        +string anonymous_id
        +string email
        +map<string, string> traits
    }

    class GetMetricsRequest {
        +string metric_names[]
        +TimeRange time_range
        +map<string, string> filters
        +string group_by[]
    }

    GetMetricsRequest "1" --> "1" TimeRange
    class MetricSeries {
        +string name
        +DataPoint points[]
        +map<string, string> metadata
    }

    MetricSeries "1" --> "*" DataPoint
    class Cell {
        +string value
        +string formatted_value
    }

    class TrackEventRequest {
        +Event event
    }

    TrackEventRequest "1" --> "1" Event
    class Chart {
        +string chart_id
        +ChartType type
        +string title
        +MetricSeries series[]
        +map<string, string> config
    }

    Chart "1" --> "1" ChartType
    Chart "1" --> "*" MetricSeries
```

---

## ⚙️ Methods

<a name="methods"></a>

This service defines **8 RPC methods**:

### TrackEvent

<a name="trackevent"></a>

TrackEvent tracks a single analytics event.

#### Method Signature

```protobuf
// Unary RPC
rpc TrackEvent(TrackEventRequest) returns (TrackEventResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.TrackEvent` |
| **Input Type** | [`TrackEventRequest`](#trackeventrequest) |
| **Output Type** | [`TrackEventResponse`](#trackeventresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: TrackEvent
    Note right of Service: TrackEventRequest
    Service-->>-Client: Response
    Note left of Client: TrackEventResponse
```

---

### BatchTrackEvents

<a name="batchtrackevents"></a>

BatchTrackEvents tracks multiple events. Uses client-side streaming to send multiple requests

#### Method Signature

```protobuf
// Client streaming RPC
rpc BatchTrackEvents(stream TrackEventRequest) returns (BatchTrackResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.BatchTrackEvents` |
| **Input Type** | [`TrackEventRequest`](#trackeventrequest) |
| **Output Type** | [`BatchTrackResponse`](#batchtrackresponse) |
| **Streaming Type** | Client Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Client Streaming
    Client->>+Service: BatchTrackEvents (stream)
    loop Stream Messages
        Client->>Service: TrackEventRequest
    end
    Service-->>-Client: BatchTrackResponse
```

---

### GetMetrics

<a name="getmetrics"></a>

GetMetrics retrieves metrics.

#### Method Signature

```protobuf
// Unary RPC
rpc GetMetrics(GetMetricsRequest) returns (GetMetricsResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetMetrics` |
| **Input Type** | [`GetMetricsRequest`](#getmetricsrequest) |
| **Output Type** | [`GetMetricsResponse`](#getmetricsresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetMetrics
    Note right of Service: GetMetricsRequest
    Service-->>-Client: Response
    Note left of Client: GetMetricsResponse
```

---

### GetReport

<a name="getreport"></a>

GetReport generates a report.

#### Method Signature

```protobuf
// Unary RPC
rpc GetReport(GetReportRequest) returns (GetReportResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetReport` |
| **Input Type** | [`GetReportRequest`](#getreportrequest) |
| **Output Type** | [`GetReportResponse`](#getreportresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetReport
    Note right of Service: GetReportRequest
    Service-->>-Client: Response
    Note left of Client: GetReportResponse
```

---

### StreamMetrics

<a name="streammetrics"></a>

StreamMetrics streams real-time metrics.

#### Method Signature

```protobuf
// Server streaming RPC
rpc StreamMetrics(StreamMetricsRequest) returns (stream MetricUpdate);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.StreamMetrics` |
| **Input Type** | [`StreamMetricsRequest`](#streammetricsrequest) |
| **Output Type** | [`MetricUpdate`](#metricupdate) |
| **Streaming Type** | Server Streaming |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Note over Client,Service: Server Streaming
    Client->>+Service: StreamMetrics
    Client->>Service: StreamMetricsRequest
    loop Stream Messages
        Service-->>Client: MetricUpdate
    end
    Service-->>-Client: End Stream
```

---

### QueryData

<a name="querydata"></a>

QueryData queries raw analytics data.

#### Method Signature

```protobuf
// Unary RPC
rpc QueryData(QueryRequest) returns (QueryResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.QueryData` |
| **Input Type** | [`QueryRequest`](#queryrequest) |
| **Output Type** | [`QueryResponse`](#queryresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: QueryData
    Note right of Service: QueryRequest
    Service-->>-Client: Response
    Note left of Client: QueryResponse
```

---

### CreateDashboard

<a name="createdashboard"></a>

CreateDashboard creates a custom dashboard.

#### Method Signature

```protobuf
// Unary RPC
rpc CreateDashboard(CreateDashboardRequest) returns (CreateDashboardResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.CreateDashboard` |
| **Input Type** | [`CreateDashboardRequest`](#createdashboardrequest) |
| **Output Type** | [`CreateDashboardResponse`](#createdashboardresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: CreateDashboard
    Note right of Service: CreateDashboardRequest
    Service-->>-Client: Response
    Note left of Client: CreateDashboardResponse
```

---

### GetDashboard

<a name="getdashboard"></a>

GetDashboard retrieves dashboard data.

#### Method Signature

```protobuf
// Unary RPC
rpc GetDashboard(GetDashboardRequest) returns (GetDashboardResponse);
```

#### Method Details

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetDashboard` |
| **Input Type** | [`GetDashboardRequest`](#getdashboardrequest) |
| **Output Type** | [`GetDashboardResponse`](#getdashboardresponse) |
| **Streaming Type** | Unary |

##### Sequence Diagram

```mermaid
%{init: {'theme':'forest'}}%
sequenceDiagram
    participant Client
    participant Service

    Client->>+Service: GetDashboard
    Note right of Service: GetDashboardRequest
    Service-->>-Client: Response
    Note left of Client: GetDashboardResponse
```

---

## 📦 Messages

<a name="messages"></a>

This service defines **40 message types**:

### DataPoint

<a name="datapoint"></a>

DataPoint represents a single data point.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.DataPoint` |
| **Field Count** | 3 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |
| 2 | `value` | double | optional | Value. |
| 3 | `labels` | map<string, string> |  | Labels. |

#### Proto Definition

```protobuf
message DataPoint {
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 1;
  // Value.
  optional double value = 2;
  // Labels.
   map<string, string> labels = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DataPoint {
        +Timestamp timestamp
        +double value
        +map<string, string> labels
    }
    DataPoint --> Timestamp
```

---

### GetReportResponse

<a name="getreportresponse"></a>

GetReportResponse returns report.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetReportResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `report` | [`Report`](#report) | optional | Report. |

#### Proto Definition

```protobuf
message GetReportResponse {
  // Report.
  optional Report report = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetReportResponse {
        +Report report
    }
    GetReportResponse --> Report
```

---

### QueryResponse

<a name="queryresponse"></a>

QueryResponse returns query results.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.QueryResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `result` | [`DataTable`](#datatable) | optional | Result table. |
| 2 | `execution_time` | [`Duration`](#duration) | optional | Execution time. |

#### Proto Definition

```protobuf
message QueryResponse {
  // Result table.
  optional DataTable result = 1;
  // Execution time.
  optional Duration execution_time = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class QueryResponse {
        +DataTable result
        +Duration execution_time
    }
    QueryResponse --> DataTable
    QueryResponse --> Duration
```

---

### GetDashboardResponse

<a name="getdashboardresponse"></a>

GetDashboardResponse returns dashboard.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetDashboardResponse` |
| **Field Count** | 2 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `dashboard` | [`Dashboard`](#dashboard) | optional | Dashboard. |
| 2 | `widget_data` | map<string, WidgetData> |  | Widget data. |

#### Proto Definition

```protobuf
message GetDashboardResponse {
  // Dashboard.
  optional Dashboard dashboard = 1;
  // Widget data.
   map<string, WidgetData> widget_data = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetDashboardResponse {
        +Dashboard dashboard
        +map<string, WidgetData> widget_data
    }
    GetDashboardResponse --> Dashboard
```

---

### DataTable

<a name="datatable"></a>

DataTable represents tabular data.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.DataTable` |
| **Field Count** | 4 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Table name. |
| 2 | `columns` | [`Column`](#column) | repeated | Columns. |
| 3 | `rows` | [`Row`](#row) | repeated | Rows. |
| 4 | `total_count` | int64 | optional | Total count Must be >= 0. |

#### Proto Definition

```protobuf
message DataTable {
  // Table name.
  optional string name = 1;
  // Columns.
  repeated Column columns = 2;
  // Rows.
  repeated Row rows = 3;
  // Total count Must be >= 0.
  optional int64 total_count = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DataTable {
        +string name
        +Column[] columns
        +Row[] rows
        +int64 total_count
    }
    DataTable "1" --> "*" Column
    DataTable "1" --> "*" Row
```

---

### Dimension

<a name="dimension"></a>

Dimension represents a data dimension.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Dimension` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Dimension name. |
| 2 | `values` | [`DimensionValue`](#dimensionvalue) | repeated | Values. |
| 3 | `total_count` | int64 | optional | Total count Must be >= 0. |

#### Proto Definition

```protobuf
message Dimension {
  // Dimension name.
  optional string name = 1;
  // Values.
  repeated DimensionValue values = 2;
  // Total count Must be >= 0.
  optional int64 total_count = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Dimension {
        +string name
        +DimensionValue[] values
        +int64 total_count
    }
    Dimension "1" --> "*" DimensionValue
```

---

### Row

<a name="row"></a>

Row represents a table row.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Row` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `cells` | [`Cell`](#cell) | repeated | Cells. |

#### Proto Definition

```protobuf
message Row {
  // Cells.
  repeated Cell cells = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Row {
        +Cell[] cells
    }
    Row "1" --> "*" Cell
```

---

### Position

<a name="position"></a>

Position represents widget position.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Position` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `x` | int32 | optional | X coordinate. |
| 2 | `y` | int32 | optional | Y coordinate. |

#### Proto Definition

```protobuf
message Position {
  // X coordinate.
  optional int32 x = 1;
  // Y coordinate.
  optional int32 y = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Position {
        +int32 x
        +int32 y
    }
```

---

### CreateDashboardRequest

<a name="createdashboardrequest"></a>

CreateDashboardRequest creates dashboard.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.CreateDashboardRequest` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `dashboard` | [`Dashboard`](#dashboard) | optional | Dashboard. |

#### Proto Definition

```protobuf
message CreateDashboardRequest {
  // Dashboard.
  optional Dashboard dashboard = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateDashboardRequest {
        +Dashboard dashboard
    }
    CreateDashboardRequest --> Dashboard
```

---

### Metric

<a name="metric"></a>

Metric represents a metric data point.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Metric` |
| **Field Count** | 6 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Metric name. |
| 2 | `type` | [`MetricType`](#metrictype) | optional | Metric type. |
| 3 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |
| 4 | `value` | double | optional | Value. |
| 5 | `tags` | map<string, string> |  | Tags/dimensions. |
| 6 | `unit` | string | optional | Unit. |

#### Proto Definition

```protobuf
message Metric {
  // Metric name.
  optional string name = 1;
  // Metric type.
  optional MetricType type = 2;
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 3;
  // Value.
  optional double value = 4;
  // Tags/dimensions.
   map<string, string> tags = 5;
  // Unit.
  optional string unit = 6;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Metric {
        +string name
        +MetricType type
        +Timestamp timestamp
        +double value
        +map<string, string> tags
        +string unit
    }
    Metric --> MetricType
    Metric --> Timestamp
```

---

### Dashboard

<a name="dashboard"></a>

Dashboard represents a custom dashboard.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Dashboard` |
| **Field Count** | 9 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `dashboard_id` | string | optional | Dashboard ID. (Must be a non-empty identifier) |
| 2 | `name` | string | optional | Name. |
| 3 | `description` | string | optional | Description. |
| 4 | `owner_id` | string | optional | Owner. (Must be a non-empty identifier) |
| 5 | `widgets` | [`Widget`](#widget) | repeated | Widgets. |
| 6 | `layout` | [`Layout`](#layout) | optional | Layout. |
| 7 | `refresh_interval` | [`Duration`](#duration) | optional | Refresh interval. |
| 8 | `created_at` | [`Timestamp`](#timestamp) | optional | Created at. (RFC 3339 timestamp format) |
| 9 | `updated_at` | [`Timestamp`](#timestamp) | optional | Updated at. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message Dashboard {
  // Dashboard ID. (Must be a non-empty identifier)
  optional string dashboard_id = 1;
  // Name.
  optional string name = 2;
  // Description.
  optional string description = 3;
  // Owner. (Must be a non-empty identifier)
  optional string owner_id = 4;
  // Widgets.
  repeated Widget widgets = 5;
  // Layout.
  optional Layout layout = 6;
  // Refresh interval.
  optional Duration refresh_interval = 7;
  // Created at. (RFC 3339 timestamp format)
  optional Timestamp created_at = 8;
  // Updated at. (RFC 3339 timestamp format)
  optional Timestamp updated_at = 9;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Dashboard {
        +string dashboard_id
        +string name
        +string description
        +string owner_id
        +Widget[] widgets
        +Layout layout
        +Duration refresh_interval
        +Timestamp created_at
        +Timestamp updated_at
    }
    Dashboard "1" --> "*" Widget
    Dashboard --> Layout
    Dashboard --> Duration
    Dashboard --> Timestamp
    Dashboard --> Timestamp
```

---

### DeviceContext

<a name="devicecontext"></a>

DeviceContext contains device information.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.DeviceContext` |
| **Field Count** | 9 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `device_type` | [`DeviceType`](#devicetype) | optional | Device type. |
| 2 | `os` | string | optional | Operating system. |
| 3 | `os_version` | string | optional | OS version. |
| 4 | `browser` | string | optional | Browser. |
| 5 | `browser_version` | string | optional | Browser version. |
| 6 | `brand` | string | optional | Device brand. |
| 7 | `model` | string | optional | Device model. |
| 8 | `screen` | [`Resolution`](#resolution) | optional | Screen resolution. |
| 9 | `user_agent` | string | optional | User agent. |

#### Proto Definition

```protobuf
message DeviceContext {
  // Device type.
  optional DeviceType device_type = 1;
  // Operating system.
  optional string os = 2;
  // OS version.
  optional string os_version = 3;
  // Browser.
  optional string browser = 4;
  // Browser version.
  optional string browser_version = 5;
  // Device brand.
  optional string brand = 6;
  // Device model.
  optional string model = 7;
  // Screen resolution.
  optional Resolution screen = 8;
  // User agent.
  optional string user_agent = 9;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DeviceContext {
        +DeviceType device_type
        +string os
        +string os_version
        +string browser
        +string browser_version
        +string brand
        +string model
        +Resolution screen
        +string user_agent
    }
    DeviceContext --> DeviceType
    DeviceContext --> Resolution
```

---

### LocationContext

<a name="locationcontext"></a>

LocationContext contains location data.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.LocationContext` |
| **Field Count** | 8 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `ip` | string | optional | IP address. |
| 2 | `country` | string | optional | Country code Must be >= 0. |
| 3 | `region` | string | optional | Region/State. |
| 4 | `city` | string | optional | City. |
| 5 | `postal_code` | string | optional | Postal code. |
| 6 | `latitude` | double | optional | Latitude. |
| 7 | `longitude` | double | optional | Longitude. |
| 8 | `timezone` | string | optional | Timezone. |

#### Proto Definition

```protobuf
message LocationContext {
  // IP address.
  optional string ip = 1;
  // Country code Must be >= 0.
  optional string country = 2;
  // Region/State.
  optional string region = 3;
  // City.
  optional string city = 4;
  // Postal code.
  optional string postal_code = 5;
  // Latitude.
  optional double latitude = 6;
  // Longitude.
  optional double longitude = 7;
  // Timezone.
  optional string timezone = 8;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class LocationContext {
        +string ip
        +string country
        +string region
        +string city
        +string postal_code
        +double latitude
        +double longitude
        +string timezone
    }
```

---

### UTMContext

<a name="utmcontext"></a>

UTMContext contains UTM parameters.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.UTMContext` |
| **Field Count** | 5 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `source` | string | optional | Source. |
| 2 | `medium` | string | optional | Medium. |
| 3 | `campaign` | string | optional | Campaign. |
| 4 | `term` | string | optional | Term. |
| 5 | `content` | string | optional | Content. |

#### Proto Definition

```protobuf
message UTMContext {
  // Source.
  optional string source = 1;
  // Medium.
  optional string medium = 2;
  // Campaign.
  optional string campaign = 3;
  // Term.
  optional string term = 4;
  // Content.
  optional string content = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UTMContext {
        +string source
        +string medium
        +string campaign
        +string term
        +string content
    }
```

---

### Widget

<a name="widget"></a>

Widget represents a dashboard widget.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Widget` |
| **Field Count** | 7 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `widget_id` | string | optional | Widget ID. (Must be a non-empty identifier) |
| 2 | `type` | [`WidgetType`](#widgettype) | optional | Widget type. |
| 3 | `title` | string | optional | Title. |
| 4 | `position` | [`Position`](#position) | optional | Position. |
| 5 | `size` | [`Size`](#size) | optional | Size. |
| 6 | `config` | map<string, string> |  | Configuration. |
| 7 | `query` | [`QueryRequest`](#queryrequest) | optional | Data query. |

#### Proto Definition

```protobuf
message Widget {
  // Widget ID. (Must be a non-empty identifier)
  optional string widget_id = 1;
  // Widget type.
  optional WidgetType type = 2;
  // Title.
  optional string title = 3;
  // Position.
  optional Position position = 4;
  // Size.
  optional Size size = 5;
  // Configuration.
   map<string, string> config = 6;
  // Data query.
  optional QueryRequest query = 7;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Widget {
        +string widget_id
        +WidgetType type
        +string title
        +Position position
        +Size size
        +map<string, string> config
        +QueryRequest query
    }
    Widget --> WidgetType
    Widget --> Position
    Widget --> Size
    Widget --> QueryRequest
```

---

### Layout

<a name="layout"></a>

Layout represents dashboard layout.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Layout` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `columns` | int32 | optional | Columns. |
| 2 | `rows` | int32 | optional | Rows. |
| 3 | `grid_size` | int32 | optional | Grid size. |

#### Proto Definition

```protobuf
message Layout {
  // Columns.
  optional int32 columns = 1;
  // Rows.
  optional int32 rows = 2;
  // Grid size.
  optional int32 grid_size = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Layout {
        +int32 columns
        +int32 rows
        +int32 grid_size
    }
```

---

### TrackEventResponse

<a name="trackeventresponse"></a>

TrackEventResponse confirms tracking.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.TrackEventResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `event_id` | string | optional | Event ID. (Must be a non-empty identifier) |
| 2 | `success` | bool | optional | Success. |

#### Proto Definition

```protobuf
message TrackEventResponse {
  // Event ID. (Must be a non-empty identifier)
  optional string event_id = 1;
  // Success.
  optional bool success = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class TrackEventResponse {
        +string event_id
        +bool success
    }
```

---

### QueryRequest

<a name="queryrequest"></a>

QueryRequest queries analytics data.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.QueryRequest` |
| **Field Count** | 4 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `query` | string | optional | SQL-like query. |
| 2 | `parameters` | map<string, string> |  | Parameters. |
| 3 | `limit` | int32 | optional | Limit Maximum value may be service-specific. |
| 4 | `offset` | int32 | optional | Offset Must be >= 0. |

#### Proto Definition

```protobuf
message QueryRequest {
  // SQL-like query.
  optional string query = 1;
  // Parameters.
   map<string, string> parameters = 2;
  // Limit Maximum value may be service-specific.
  optional int32 limit = 3;
  // Offset Must be >= 0.
  optional int32 offset = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class QueryRequest {
        +string query
        +map<string, string> parameters
        +int32 limit
        +int32 offset
    }
```

---

### MetricSummary

<a name="metricsummary"></a>

MetricSummary contains aggregated metric data.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.MetricSummary` |
| **Field Count** | 6 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Metric name. |
| 2 | `current_value` | double | optional | Current value. |
| 3 | `previous_value` | double | optional | Previous value. |
| 4 | `change_percent` | double | optional | Change percentage Range: 0-100. |
| 5 | `trend` | [`Trend`](#trend) | optional | Trend. |
| 6 | `aggregation` | [`AggregationType`](#aggregationtype) | optional | Aggregation type. |

#### Proto Definition

```protobuf
message MetricSummary {
  // Metric name.
  optional string name = 1;
  // Current value.
  optional double current_value = 2;
  // Previous value.
  optional double previous_value = 3;
  // Change percentage Range: 0-100.
  optional double change_percent = 4;
  // Trend.
  optional Trend trend = 5;
  // Aggregation type.
  optional AggregationType aggregation = 6;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class MetricSummary {
        +string name
        +double current_value
        +double previous_value
        +double change_percent
        +Trend trend
        +AggregationType aggregation
    }
    MetricSummary --> Trend
    MetricSummary --> AggregationType
```

---

### DimensionValue

<a name="dimensionvalue"></a>

DimensionValue represents dimension breakdown.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.DimensionValue` |
| **Field Count** | 4 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `value` | string | optional | Value. |
| 2 | `count` | int64 | optional | Count Must be >= 0. |
| 3 | `percentage` | double | optional | Percentage Range: 0-100. |
| 4 | `metrics` | map<string, double> |  | Metrics. |

#### Proto Definition

```protobuf
message DimensionValue {
  // Value.
  optional string value = 1;
  // Count Must be >= 0.
  optional int64 count = 2;
  // Percentage Range: 0-100.
  optional double percentage = 3;
  // Metrics.
   map<string, double> metrics = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class DimensionValue {
        +string value
        +int64 count
        +double percentage
        +map<string, double> metrics
    }
```

---

### GetMetricsResponse

<a name="getmetricsresponse"></a>

GetMetricsResponse returns metrics.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetMetricsResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `metrics` | [`MetricSeries`](#metricseries) | repeated | Metrics. |

#### Proto Definition

```protobuf
message GetMetricsResponse {
  // Metrics.
  repeated MetricSeries metrics = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetMetricsResponse {
        +MetricSeries[] metrics
    }
    GetMetricsResponse "1" --> "*" MetricSeries
```

---

### StreamMetricsRequest

<a name="streammetricsrequest"></a>

StreamMetricsRequest subscribes to metrics.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.StreamMetricsRequest` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `metric_names` | string | repeated | Metric names. |
| 2 | `interval` | [`Duration`](#duration) | optional | Update interval. |

#### Proto Definition

```protobuf
message StreamMetricsRequest {
  // Metric names.
  repeated string metric_names = 1;
  // Update interval.
  optional Duration interval = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class StreamMetricsRequest {
        +string[] metric_names
        +Duration interval
    }
    StreamMetricsRequest --> Duration
```

---

### MetricUpdate

<a name="metricupdate"></a>

MetricUpdate streams metric updates.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.MetricUpdate` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `metric` | [`Metric`](#metric) | optional | Metric. |
| 2 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message MetricUpdate {
  // Metric.
  optional Metric metric = 1;
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class MetricUpdate {
        +Metric metric
        +Timestamp timestamp
    }
    MetricUpdate --> Metric
    MetricUpdate --> Timestamp
```

---

### CreateDashboardResponse

<a name="createdashboardresponse"></a>

CreateDashboardResponse confirms creation.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.CreateDashboardResponse` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `dashboard` | [`Dashboard`](#dashboard) | optional | Created dashboard. |

#### Proto Definition

```protobuf
message CreateDashboardResponse {
  // Created dashboard.
  optional Dashboard dashboard = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class CreateDashboardResponse {
        +Dashboard dashboard
    }
    CreateDashboardResponse --> Dashboard
```

---

### TimeRange

<a name="timerange"></a>

TimeRange represents a time period.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.TimeRange` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `start` | [`Timestamp`](#timestamp) | optional | Start time. |
| 2 | `end` | [`Timestamp`](#timestamp) | optional | End time. |
| 3 | `granularity` | [`TimeGranularity`](#timegranularity) | optional | Granularity. |

#### Proto Definition

```protobuf
message TimeRange {
  // Start time.
  optional Timestamp start = 1;
  // End time.
  optional Timestamp end = 2;
  // Granularity.
  optional TimeGranularity granularity = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class TimeRange {
        +Timestamp start
        +Timestamp end
        +TimeGranularity granularity
    }
    TimeRange --> Timestamp
    TimeRange --> Timestamp
    TimeRange --> TimeGranularity
```

---

### Resolution

<a name="resolution"></a>

Resolution represents screen resolution.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Resolution` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `width` | int32 | optional | Width. |
| 2 | `height` | int32 | optional | Height. |

#### Proto Definition

```protobuf
message Resolution {
  // Width.
  optional int32 width = 1;
  // Height.
  optional int32 height = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Resolution {
        +int32 width
        +int32 height
    }
```

---

### Size

<a name="size"></a>

Size represents widget size.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Size` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `width` | int32 | optional | Width. |
| 2 | `height` | int32 | optional | Height. |

#### Proto Definition

```protobuf
message Size {
  // Width.
  optional int32 width = 1;
  // Height.
  optional int32 height = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Size {
        +int32 width
        +int32 height
    }
```

---

### SessionContext

<a name="sessioncontext"></a>

SessionContext contains session information.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.SessionContext` |
| **Field Count** | 5 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `session_id` | string | optional | Session ID. (Must be a non-empty identifier) |
| 2 | `started_at` | [`Timestamp`](#timestamp) | optional | Session start time. (RFC 3339 timestamp format) |
| 3 | `duration` | [`Duration`](#duration) | optional | Session duration. |
| 4 | `page_views` | int32 | optional | Page views in session. |
| 5 | `event_count` | int32 | optional | Events in session Must be >= 0. |

#### Proto Definition

```protobuf
message SessionContext {
  // Session ID. (Must be a non-empty identifier)
  optional string session_id = 1;
  // Session start time. (RFC 3339 timestamp format)
  optional Timestamp started_at = 2;
  // Session duration.
  optional Duration duration = 3;
  // Page views in session.
  optional int32 page_views = 4;
  // Events in session Must be >= 0.
  optional int32 event_count = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class SessionContext {
        +string session_id
        +Timestamp started_at
        +Duration duration
        +int32 page_views
        +int32 event_count
    }
    SessionContext --> Timestamp
    SessionContext --> Duration
```

---

### BatchTrackResponse

<a name="batchtrackresponse"></a>

BatchTrackResponse confirms batch tracking.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.BatchTrackResponse` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `events_tracked` | int32 | optional | Events tracked. |
| 2 | `failed_count` | int32 | optional | Failed count Must be >= 0. |

#### Proto Definition

```protobuf
message BatchTrackResponse {
  // Events tracked.
  optional int32 events_tracked = 1;
  // Failed count Must be >= 0.
  optional int32 failed_count = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class BatchTrackResponse {
        +int32 events_tracked
        +int32 failed_count
    }
```

---

### GetDashboardRequest

<a name="getdashboardrequest"></a>

GetDashboardRequest retrieves dashboard.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetDashboardRequest` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `dashboard_id` | string | optional | Dashboard ID. (Must be a non-empty identifier) |

#### Proto Definition

```protobuf
message GetDashboardRequest {
  // Dashboard ID. (Must be a non-empty identifier)
  optional string dashboard_id = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetDashboardRequest {
        +string dashboard_id
    }
```

---

### Event

<a name="event"></a>

Event represents an analytics event.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Event` |
| **Field Count** | 14 |
| **Nested Types** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `event_id` | string | optional | Event ID. (Must be a non-empty identifier) |
| 2 | `event_type` | [`EventType`](#eventtype) | optional | Event type. |
| 3 | `event_name` | string | optional | Event name. |
| 4 | `timestamp` | [`Timestamp`](#timestamp) | optional | Timestamp. (RFC 3339 timestamp format) |
| 5 | `user` | [`UserContext`](#usercontext) | optional | User information. |
| 6 | `session` | [`SessionContext`](#sessioncontext) | optional | Session information. |
| 7 | `device` | [`DeviceContext`](#devicecontext) | optional | Device information. |
| 8 | `location` | [`LocationContext`](#locationcontext) | optional | Location information. |
| 9 | `properties` | map<string, string> |  | Event properties. |
| 10 | `value` | double | optional | Event value (for revenue tracking). |
| 11 | `currency` | string | optional | Currency (for revenue events). |
| 12 | `utm` | [`UTMContext`](#utmcontext) | optional | UTM parameters. |
| 13 | `referrer` | string | optional | Referrer. |
| 14 | `dimensions` | map<string, string> |  | Custom dimensions. |

#### Proto Definition

```protobuf
message Event {
  // Event ID. (Must be a non-empty identifier)
  optional string event_id = 1;
  // Event type.
  optional EventType event_type = 2;
  // Event name.
  optional string event_name = 3;
  // Timestamp. (RFC 3339 timestamp format)
  optional Timestamp timestamp = 4;
  // User information.
  optional UserContext user = 5;
  // Session information.
  optional SessionContext session = 6;
  // Device information.
  optional DeviceContext device = 7;
  // Location information.
  optional LocationContext location = 8;
  // Event properties.
   map<string, string> properties = 9;
  // Event value (for revenue tracking).
  optional double value = 10;
  // Currency (for revenue events).
  optional string currency = 11;
  // UTM parameters.
  optional UTMContext utm = 12;
  // Referrer.
  optional string referrer = 13;
  // Custom dimensions.
   map<string, string> dimensions = 14;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Event {
        +string event_id
        +EventType event_type
        +string event_name
        +Timestamp timestamp
        +UserContext user
        +SessionContext session
        +DeviceContext device
        +LocationContext location
        +map<string, string> properties
        +double value
        +string currency
        +UTMContext utm
        +string referrer
        +map<string, string> dimensions
    }
    Event --> EventType
    Event --> Timestamp
    Event --> UserContext
    Event --> SessionContext
    Event --> DeviceContext
    Event --> LocationContext
    Event --> UTMContext
```

---

### Report

<a name="report"></a>

Report represents an analytics report.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Report` |
| **Field Count** | 9 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `report_id` | string | optional | Report ID. (Must be a non-empty identifier) |
| 2 | `name` | string | optional | Report name. |
| 3 | `type` | [`ReportType`](#reporttype) | optional | Report type. |
| 4 | `time_range` | [`TimeRange`](#timerange) | optional | Time range. |
| 5 | `metrics` | [`MetricSummary`](#metricsummary) | repeated | Metrics. |
| 6 | `dimensions` | [`Dimension`](#dimension) | repeated | Dimensions. |
| 7 | `charts` | [`Chart`](#chart) | repeated | Charts. |
| 8 | `tables` | [`DataTable`](#datatable) | repeated | Tables. |
| 9 | `generated_at` | [`Timestamp`](#timestamp) | optional | Generated at. (RFC 3339 timestamp format) |

#### Proto Definition

```protobuf
message Report {
  // Report ID. (Must be a non-empty identifier)
  optional string report_id = 1;
  // Report name.
  optional string name = 2;
  // Report type.
  optional ReportType type = 3;
  // Time range.
  optional TimeRange time_range = 4;
  // Metrics.
  repeated MetricSummary metrics = 5;
  // Dimensions.
  repeated Dimension dimensions = 6;
  // Charts.
  repeated Chart charts = 7;
  // Tables.
  repeated DataTable tables = 8;
  // Generated at. (RFC 3339 timestamp format)
  optional Timestamp generated_at = 9;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Report {
        +string report_id
        +string name
        +ReportType type
        +TimeRange time_range
        +MetricSummary[] metrics
        +Dimension[] dimensions
        +Chart[] charts
        +DataTable[] tables
        +Timestamp generated_at
    }
    Report --> ReportType
    Report --> TimeRange
    Report "1" --> "*" MetricSummary
    Report "1" --> "*" Dimension
    Report "1" --> "*" Chart
    Report "1" --> "*" DataTable
    Report --> Timestamp
```

---

### Column

<a name="column"></a>

Column represents a table column.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Column` |
| **Field Count** | 3 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Column name. |
| 2 | `type` | [`DataType`](#datatype) | optional | Data type. |
| 3 | `format` | string | optional | Format. |

#### Proto Definition

```protobuf
message Column {
  // Column name.
  optional string name = 1;
  // Data type.
  optional DataType type = 2;
  // Format.
  optional string format = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Column {
        +string name
        +DataType type
        +string format
    }
    Column --> DataType
```

---

### GetReportRequest

<a name="getreportrequest"></a>

GetReportRequest generates a report.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetReportRequest` |
| **Field Count** | 5 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `type` | [`ReportType`](#reporttype) | optional | Report type. |
| 2 | `time_range` | [`TimeRange`](#timerange) | optional | Time range. |
| 3 | `filters` | map<string, string> |  | Filters. |
| 4 | `metrics` | string | repeated | Metrics to include. |
| 5 | `dimensions` | string | repeated | Dimensions to group by. |

#### Proto Definition

```protobuf
message GetReportRequest {
  // Report type.
  optional ReportType type = 1;
  // Time range.
  optional TimeRange time_range = 2;
  // Filters.
   map<string, string> filters = 3;
  // Metrics to include.
  repeated string metrics = 4;
  // Dimensions to group by.
  repeated string dimensions = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetReportRequest {
        +ReportType type
        +TimeRange time_range
        +map<string, string> filters
        +string[] metrics
        +string[] dimensions
    }
    GetReportRequest --> ReportType
    GetReportRequest --> TimeRange
```

---

### UserContext

<a name="usercontext"></a>

UserContext contains user information.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.UserContext` |
| **Field Count** | 4 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `user_id` | string | optional | User ID. (Must be a non-empty identifier) |
| 2 | `anonymous_id` | string | optional | Anonymous ID. (Must be a non-empty identifier) |
| 3 | `email` | string | optional | Email. (Must be a valid email address format) |
| 4 | `traits` | map<string, string> |  | User traits. |

#### Proto Definition

```protobuf
message UserContext {
  // User ID. (Must be a non-empty identifier)
  optional string user_id = 1;
  // Anonymous ID. (Must be a non-empty identifier)
  optional string anonymous_id = 2;
  // Email. (Must be a valid email address format)
  optional string email = 3;
  // User traits.
   map<string, string> traits = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class UserContext {
        +string user_id
        +string anonymous_id
        +string email
        +map<string, string> traits
    }
```

---

### GetMetricsRequest

<a name="getmetricsrequest"></a>

GetMetricsRequest retrieves metrics.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.GetMetricsRequest` |
| **Field Count** | 4 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `metric_names` | string | repeated | Metric names. |
| 2 | `time_range` | [`TimeRange`](#timerange) | optional | Time range. |
| 3 | `filters` | map<string, string> |  | Filters. |
| 4 | `group_by` | string | repeated | Group by dimensions. |

#### Proto Definition

```protobuf
message GetMetricsRequest {
  // Metric names.
  repeated string metric_names = 1;
  // Time range.
  optional TimeRange time_range = 2;
  // Filters.
   map<string, string> filters = 3;
  // Group by dimensions.
  repeated string group_by = 4;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class GetMetricsRequest {
        +string[] metric_names
        +TimeRange time_range
        +map<string, string> filters
        +string[] group_by
    }
    GetMetricsRequest --> TimeRange
```

---

### MetricSeries

<a name="metricseries"></a>

MetricSeries represents time series data.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.MetricSeries` |
| **Field Count** | 3 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `name` | string | optional | Series name. |
| 2 | `points` | [`DataPoint`](#datapoint) | repeated | Data points. |
| 3 | `metadata` | map<string, string> |  | Metadata. |

#### Proto Definition

```protobuf
message MetricSeries {
  // Series name.
  optional string name = 1;
  // Data points.
  repeated DataPoint points = 2;
  // Metadata.
   map<string, string> metadata = 3;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class MetricSeries {
        +string name
        +DataPoint[] points
        +map<string, string> metadata
    }
    MetricSeries "1" --> "*" DataPoint
```

---

### Cell

<a name="cell"></a>

Cell represents a table cell.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Cell` |
| **Field Count** | 2 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `value` | string | optional | Value (as string). |
| 2 | `formatted_value` | string | optional | Formatted value. |

#### Proto Definition

```protobuf
message Cell {
  // Value (as string).
  optional string value = 1;
  // Formatted value.
  optional string formatted_value = 2;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Cell {
        +string value
        +string formatted_value
    }
```

---

### TrackEventRequest

<a name="trackeventrequest"></a>

TrackEventRequest tracks an event.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.TrackEventRequest` |
| **Field Count** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `event` | [`Event`](#event) | optional | Event. |

#### Proto Definition

```protobuf
message TrackEventRequest {
  // Event.
  optional Event event = 1;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class TrackEventRequest {
        +Event event
    }
    TrackEventRequest --> Event
```

---

### Chart

<a name="chart"></a>

Chart represents a chart visualization.

| Attribute | Value |
|-----------|-------|
| **Full Name** | `analytics.v1.Chart` |
| **Field Count** | 5 |
| **Nested Types** | 1 |

#### Fields

| # | Name | Type | Label | Description |
|---|------|------|-------|-------------|
| 1 | `chart_id` | string | optional | Chart ID. (Must be a non-empty identifier) |
| 2 | `type` | [`ChartType`](#charttype) | optional | Chart type. |
| 3 | `title` | string | optional | Title. |
| 4 | `series` | [`MetricSeries`](#metricseries) | repeated | Data series. |
| 5 | `config` | map<string, string> |  | Configuration. |

#### Proto Definition

```protobuf
message Chart {
  // Chart ID. (Must be a non-empty identifier)
  optional string chart_id = 1;
  // Chart type.
  optional ChartType type = 2;
  // Title.
  optional string title = 3;
  // Data series.
  repeated MetricSeries series = 4;
  // Configuration.
   map<string, string> config = 5;
}
```

##### Message Structure

```mermaid
%{init: {'theme':'forest'}}%
classDiagram
    class Chart {
        +string chart_id
        +ChartType type
        +string title
        +MetricSeries[] series
        +map<string, string> config
    }
    Chart --> ChartType
    Chart "1" --> "*" MetricSeries
```

---

## 🔢 Enumerations

<a name="enumerations"></a>

This service defines **10 enumeration types**:

### AggregationType

<a name="aggregationtype"></a>

AggregationType represents aggregation methods.

| Value | Number | Description |
|-------|--------|-------------|
| `AGGREGATION_TYPE_UNSPECIFIED` | 0 | AGGREGATION_TYPE_UNSPECIFIED value. |
| `AGGREGATION_TYPE_SUM` | 1 | AGGREGATION_TYPE_SUM value. |
| `AGGREGATION_TYPE_AVG` | 2 | AGGREGATION_TYPE_AVG value. |
| `AGGREGATION_TYPE_MIN` | 3 | AGGREGATION_TYPE_MIN value. |
| `AGGREGATION_TYPE_MAX` | 4 | AGGREGATION_TYPE_MAX value. |
| `AGGREGATION_TYPE_COUNT` | 5 | AGGREGATION_TYPE_COUNT value. |
| `AGGREGATION_TYPE_PERCENTILE` | 6 | AGGREGATION_TYPE_PERCENTILE value. |

#### Proto Definition

```protobuf
enum AggregationType {
  // AGGREGATION_TYPE_UNSPECIFIED value.
  AGGREGATION_TYPE_UNSPECIFIED = 0;
  // AGGREGATION_TYPE_SUM value.
  AGGREGATION_TYPE_SUM = 1;
  // AGGREGATION_TYPE_AVG value.
  AGGREGATION_TYPE_AVG = 2;
  // AGGREGATION_TYPE_MIN value.
  AGGREGATION_TYPE_MIN = 3;
  // AGGREGATION_TYPE_MAX value.
  AGGREGATION_TYPE_MAX = 4;
  // AGGREGATION_TYPE_COUNT value.
  AGGREGATION_TYPE_COUNT = 5;
  // AGGREGATION_TYPE_PERCENTILE value.
  AGGREGATION_TYPE_PERCENTILE = 6;
}
```

---

### DeviceType

<a name="devicetype"></a>

DeviceType represents device types.

| Value | Number | Description |
|-------|--------|-------------|
| `DEVICE_TYPE_UNSPECIFIED` | 0 | DEVICE_TYPE_UNSPECIFIED value. |
| `DEVICE_TYPE_DESKTOP` | 1 | DEVICE_TYPE_DESKTOP value. |
| `DEVICE_TYPE_MOBILE` | 2 | DEVICE_TYPE_MOBILE value. |
| `DEVICE_TYPE_TABLET` | 3 | DEVICE_TYPE_TABLET value. |
| `DEVICE_TYPE_TV` | 4 | DEVICE_TYPE_TV value. |
| `DEVICE_TYPE_WEARABLE` | 5 | DEVICE_TYPE_WEARABLE value. |

#### Proto Definition

```protobuf
enum DeviceType {
  // DEVICE_TYPE_UNSPECIFIED value.
  DEVICE_TYPE_UNSPECIFIED = 0;
  // DEVICE_TYPE_DESKTOP value.
  DEVICE_TYPE_DESKTOP = 1;
  // DEVICE_TYPE_MOBILE value.
  DEVICE_TYPE_MOBILE = 2;
  // DEVICE_TYPE_TABLET value.
  DEVICE_TYPE_TABLET = 3;
  // DEVICE_TYPE_TV value.
  DEVICE_TYPE_TV = 4;
  // DEVICE_TYPE_WEARABLE value.
  DEVICE_TYPE_WEARABLE = 5;
}
```

---

### ReportType

<a name="reporttype"></a>

ReportType represents report types.

| Value | Number | Description |
|-------|--------|-------------|
| `REPORT_TYPE_UNSPECIFIED` | 0 | REPORT_TYPE_UNSPECIFIED value. |
| `REPORT_TYPE_OVERVIEW` | 1 | REPORT_TYPE_OVERVIEW value. |
| `REPORT_TYPE_USER_ACTIVITY` | 2 | REPORT_TYPE_USER_ACTIVITY value. |
| `REPORT_TYPE_CONVERSION` | 3 | REPORT_TYPE_CONVERSION value. |
| `REPORT_TYPE_REVENUE` | 4 | REPORT_TYPE_REVENUE value. |
| `REPORT_TYPE_RETENTION` | 5 | REPORT_TYPE_RETENTION value. |
| `REPORT_TYPE_FUNNEL` | 6 | REPORT_TYPE_FUNNEL value. |
| `REPORT_TYPE_COHORT` | 7 | REPORT_TYPE_COHORT value. |
| `REPORT_TYPE_CUSTOM` | 100 | REPORT_TYPE_CUSTOM value. |

#### Proto Definition

```protobuf
enum ReportType {
  // REPORT_TYPE_UNSPECIFIED value.
  REPORT_TYPE_UNSPECIFIED = 0;
  // REPORT_TYPE_OVERVIEW value.
  REPORT_TYPE_OVERVIEW = 1;
  // REPORT_TYPE_USER_ACTIVITY value.
  REPORT_TYPE_USER_ACTIVITY = 2;
  // REPORT_TYPE_CONVERSION value.
  REPORT_TYPE_CONVERSION = 3;
  // REPORT_TYPE_REVENUE value.
  REPORT_TYPE_REVENUE = 4;
  // REPORT_TYPE_RETENTION value.
  REPORT_TYPE_RETENTION = 5;
  // REPORT_TYPE_FUNNEL value.
  REPORT_TYPE_FUNNEL = 6;
  // REPORT_TYPE_COHORT value.
  REPORT_TYPE_COHORT = 7;
  // REPORT_TYPE_CUSTOM value.
  REPORT_TYPE_CUSTOM = 100;
}
```

---

### Trend

<a name="trend"></a>

Trend represents data trends.

| Value | Number | Description |
|-------|--------|-------------|
| `TREND_UNSPECIFIED` | 0 | TREND_UNSPECIFIED value. |
| `TREND_UP` | 1 | TREND_UP value. |
| `TREND_DOWN` | 2 | TREND_DOWN value. |
| `TREND_STABLE` | 3 | TREND_STABLE value. |

#### Proto Definition

```protobuf
enum Trend {
  // TREND_UNSPECIFIED value.
  TREND_UNSPECIFIED = 0;
  // TREND_UP value.
  TREND_UP = 1;
  // TREND_DOWN value.
  TREND_DOWN = 2;
  // TREND_STABLE value.
  TREND_STABLE = 3;
}
```

---

### ChartType

<a name="charttype"></a>

ChartType represents chart types.

| Value | Number | Description |
|-------|--------|-------------|
| `CHART_TYPE_UNSPECIFIED` | 0 | CHART_TYPE_UNSPECIFIED value. |
| `CHART_TYPE_LINE` | 1 | CHART_TYPE_LINE value. |
| `CHART_TYPE_BAR` | 2 | CHART_TYPE_BAR value. |
| `CHART_TYPE_PIE` | 3 | CHART_TYPE_PIE value. |
| `CHART_TYPE_AREA` | 4 | CHART_TYPE_AREA value. |
| `CHART_TYPE_SCATTER` | 5 | CHART_TYPE_SCATTER value. |
| `CHART_TYPE_HEATMAP` | 6 | CHART_TYPE_HEATMAP value. |
| `CHART_TYPE_FUNNEL` | 7 | CHART_TYPE_FUNNEL value. |

#### Proto Definition

```protobuf
enum ChartType {
  // CHART_TYPE_UNSPECIFIED value.
  CHART_TYPE_UNSPECIFIED = 0;
  // CHART_TYPE_LINE value.
  CHART_TYPE_LINE = 1;
  // CHART_TYPE_BAR value.
  CHART_TYPE_BAR = 2;
  // CHART_TYPE_PIE value.
  CHART_TYPE_PIE = 3;
  // CHART_TYPE_AREA value.
  CHART_TYPE_AREA = 4;
  // CHART_TYPE_SCATTER value.
  CHART_TYPE_SCATTER = 5;
  // CHART_TYPE_HEATMAP value.
  CHART_TYPE_HEATMAP = 6;
  // CHART_TYPE_FUNNEL value.
  CHART_TYPE_FUNNEL = 7;
}
```

---

### EventType

<a name="eventtype"></a>

EventType represents tracked event types.

| Value | Number | Description |
|-------|--------|-------------|
| `EVENT_TYPE_UNSPECIFIED` | 0 | EVENT_TYPE_UNSPECIFIED value. |
| `EVENT_TYPE_PAGE_VIEW` | 1 | EVENT_TYPE_PAGE_VIEW value. |
| `EVENT_TYPE_CLICK` | 2 | EVENT_TYPE_CLICK value. |
| `EVENT_TYPE_CONVERSION` | 3 | EVENT_TYPE_CONVERSION value. |
| `EVENT_TYPE_PURCHASE` | 4 | EVENT_TYPE_PURCHASE value. |
| `EVENT_TYPE_SIGNUP` | 5 | EVENT_TYPE_SIGNUP value. |
| `EVENT_TYPE_LOGIN` | 6 | EVENT_TYPE_LOGIN value. |
| `EVENT_TYPE_LOGOUT` | 7 | EVENT_TYPE_LOGOUT value. |
| `EVENT_TYPE_SEARCH` | 8 | EVENT_TYPE_SEARCH value. |
| `EVENT_TYPE_SHARE` | 9 | EVENT_TYPE_SHARE value. |
| `EVENT_TYPE_CUSTOM` | 100 | EVENT_TYPE_CUSTOM value. |

#### Proto Definition

```protobuf
enum EventType {
  // EVENT_TYPE_UNSPECIFIED value.
  EVENT_TYPE_UNSPECIFIED = 0;
  // EVENT_TYPE_PAGE_VIEW value.
  EVENT_TYPE_PAGE_VIEW = 1;
  // EVENT_TYPE_CLICK value.
  EVENT_TYPE_CLICK = 2;
  // EVENT_TYPE_CONVERSION value.
  EVENT_TYPE_CONVERSION = 3;
  // EVENT_TYPE_PURCHASE value.
  EVENT_TYPE_PURCHASE = 4;
  // EVENT_TYPE_SIGNUP value.
  EVENT_TYPE_SIGNUP = 5;
  // EVENT_TYPE_LOGIN value.
  EVENT_TYPE_LOGIN = 6;
  // EVENT_TYPE_LOGOUT value.
  EVENT_TYPE_LOGOUT = 7;
  // EVENT_TYPE_SEARCH value.
  EVENT_TYPE_SEARCH = 8;
  // EVENT_TYPE_SHARE value.
  EVENT_TYPE_SHARE = 9;
  // EVENT_TYPE_CUSTOM value.
  EVENT_TYPE_CUSTOM = 100;
}
```

---

### TimeGranularity

<a name="timegranularity"></a>

TimeGranularity represents time bucket sizes.

| Value | Number | Description |
|-------|--------|-------------|
| `TIME_GRANULARITY_UNSPECIFIED` | 0 | TIME_GRANULARITY_UNSPECIFIED value. |
| `TIME_GRANULARITY_MINUTE` | 1 | TIME_GRANULARITY_MINUTE value. |
| `TIME_GRANULARITY_HOUR` | 2 | TIME_GRANULARITY_HOUR value. |
| `TIME_GRANULARITY_DAY` | 3 | TIME_GRANULARITY_DAY value. |
| `TIME_GRANULARITY_WEEK` | 4 | TIME_GRANULARITY_WEEK value. |
| `TIME_GRANULARITY_MONTH` | 5 | TIME_GRANULARITY_MONTH value. |
| `TIME_GRANULARITY_YEAR` | 6 | TIME_GRANULARITY_YEAR value. |

#### Proto Definition

```protobuf
enum TimeGranularity {
  // TIME_GRANULARITY_UNSPECIFIED value.
  TIME_GRANULARITY_UNSPECIFIED = 0;
  // TIME_GRANULARITY_MINUTE value.
  TIME_GRANULARITY_MINUTE = 1;
  // TIME_GRANULARITY_HOUR value.
  TIME_GRANULARITY_HOUR = 2;
  // TIME_GRANULARITY_DAY value.
  TIME_GRANULARITY_DAY = 3;
  // TIME_GRANULARITY_WEEK value.
  TIME_GRANULARITY_WEEK = 4;
  // TIME_GRANULARITY_MONTH value.
  TIME_GRANULARITY_MONTH = 5;
  // TIME_GRANULARITY_YEAR value.
  TIME_GRANULARITY_YEAR = 6;
}
```

---

### DataType

<a name="datatype"></a>

DataType represents data types.

| Value | Number | Description |
|-------|--------|-------------|
| `DATA_TYPE_UNSPECIFIED` | 0 | DATA_TYPE_UNSPECIFIED value. |
| `DATA_TYPE_STRING` | 1 | DATA_TYPE_STRING value. |
| `DATA_TYPE_NUMBER` | 2 | DATA_TYPE_NUMBER value. |
| `DATA_TYPE_BOOLEAN` | 3 | DATA_TYPE_BOOLEAN value. |
| `DATA_TYPE_TIMESTAMP` | 4 | DATA_TYPE_TIMESTAMP value. |
| `DATA_TYPE_DURATION` | 5 | DATA_TYPE_DURATION value. |

#### Proto Definition

```protobuf
enum DataType {
  // DATA_TYPE_UNSPECIFIED value.
  DATA_TYPE_UNSPECIFIED = 0;
  // DATA_TYPE_STRING value.
  DATA_TYPE_STRING = 1;
  // DATA_TYPE_NUMBER value.
  DATA_TYPE_NUMBER = 2;
  // DATA_TYPE_BOOLEAN value.
  DATA_TYPE_BOOLEAN = 3;
  // DATA_TYPE_TIMESTAMP value.
  DATA_TYPE_TIMESTAMP = 4;
  // DATA_TYPE_DURATION value.
  DATA_TYPE_DURATION = 5;
}
```

---

### WidgetType

<a name="widgettype"></a>

WidgetType represents widget types.

| Value | Number | Description |
|-------|--------|-------------|
| `WIDGET_TYPE_UNSPECIFIED` | 0 | WIDGET_TYPE_UNSPECIFIED value. |
| `WIDGET_TYPE_METRIC` | 1 | WIDGET_TYPE_METRIC value. |
| `WIDGET_TYPE_CHART` | 2 | WIDGET_TYPE_CHART value. |
| `WIDGET_TYPE_TABLE` | 3 | WIDGET_TYPE_TABLE value. |
| `WIDGET_TYPE_TEXT` | 4 | WIDGET_TYPE_TEXT value. |
| `WIDGET_TYPE_CUSTOM` | 100 | WIDGET_TYPE_CUSTOM value. |

#### Proto Definition

```protobuf
enum WidgetType {
  // WIDGET_TYPE_UNSPECIFIED value.
  WIDGET_TYPE_UNSPECIFIED = 0;
  // WIDGET_TYPE_METRIC value.
  WIDGET_TYPE_METRIC = 1;
  // WIDGET_TYPE_CHART value.
  WIDGET_TYPE_CHART = 2;
  // WIDGET_TYPE_TABLE value.
  WIDGET_TYPE_TABLE = 3;
  // WIDGET_TYPE_TEXT value.
  WIDGET_TYPE_TEXT = 4;
  // WIDGET_TYPE_CUSTOM value.
  WIDGET_TYPE_CUSTOM = 100;
}
```

---

### MetricType

<a name="metrictype"></a>

MetricType represents metric types.

| Value | Number | Description |
|-------|--------|-------------|
| `METRIC_TYPE_UNSPECIFIED` | 0 | METRIC_TYPE_UNSPECIFIED value. |
| `METRIC_TYPE_COUNTER` | 1 | METRIC_TYPE_COUNTER value. |
| `METRIC_TYPE_GAUGE` | 2 | METRIC_TYPE_GAUGE value. |
| `METRIC_TYPE_HISTOGRAM` | 3 | METRIC_TYPE_HISTOGRAM value. |
| `METRIC_TYPE_SUMMARY` | 4 | METRIC_TYPE_SUMMARY value. |

#### Proto Definition

```protobuf
enum MetricType {
  // METRIC_TYPE_UNSPECIFIED value.
  METRIC_TYPE_UNSPECIFIED = 0;
  // METRIC_TYPE_COUNTER value.
  METRIC_TYPE_COUNTER = 1;
  // METRIC_TYPE_GAUGE value.
  METRIC_TYPE_GAUGE = 2;
  // METRIC_TYPE_HISTOGRAM value.
  METRIC_TYPE_HISTOGRAM = 3;
  // METRIC_TYPE_SUMMARY value.
  METRIC_TYPE_SUMMARY = 4;
}
```

---

## 🗄️ Data Model (ERD)

<a name="erd"></a>

Entity-Relationship diagram showing the data model.

```mermaid
%{init: {'theme':'forest'}}%
erDiagram
    DataPoint {
        Timestamp timestamp
        double value
        map<string, string> labels
    }

    GetReportResponse {
        Report report
    }

    GetReportResponse ||--|| Report : has
    QueryResponse {
        DataTable result
        Duration execution_time
    }

    QueryResponse ||--|| DataTable : has
    GetDashboardResponse {
        Dashboard dashboard
        map<string, WidgetData> widget_data
    }

    GetDashboardResponse ||--|| Dashboard : has
    DataTable {
        string name
        Column columns
        Row rows
        int64 total_count
    }

    DataTable ||--o{ Column : has
    DataTable ||--o{ Row : has
    Dimension {
        string name
        DimensionValue values
        int64 total_count
    }

    Dimension ||--o{ DimensionValue : has
    Row {
        Cell cells
    }

    Row ||--o{ Cell : has
    Position {
        int32 x
        int32 y
    }

    CreateDashboardRequest {
        Dashboard dashboard
    }

    CreateDashboardRequest ||--|| Dashboard : has
    Metric {
        string name
        MetricType type
        Timestamp timestamp
        double value
        map<string, string> tags
        string unit
    }

    Metric ||--|| MetricType : has
    Dashboard {
        string dashboard_id
        string name
        string description
        string owner_id
        Widget widgets
        Layout layout
        Duration refresh_interval
        Timestamp created_at
        Timestamp updated_at
    }

    Dashboard ||--o{ Widget : has
    Dashboard ||--|| Layout : has
    DeviceContext {
        DeviceType device_type
        string os
        string os_version
        string browser
        string browser_version
        string brand
        string model
        Resolution screen
        string user_agent
    }

    DeviceContext ||--|| DeviceType : has
    DeviceContext ||--|| Resolution : has
    LocationContext {
        string ip
        string country
        string region
        string city
        string postal_code
        double latitude
        double longitude
        string timezone
    }

    UTMContext {
        string source
        string medium
        string campaign
        string term
        string content
    }

    Widget {
        string widget_id
        WidgetType type
        string title
        Position position
        Size size
        map<string, string> config
        QueryRequest query
    }

    Widget ||--|| WidgetType : has
    Widget ||--|| Position : has
    Widget ||--|| Size : has
    Widget ||--|| QueryRequest : has
    Layout {
        int32 columns
        int32 rows
        int32 grid_size
    }

    TrackEventResponse {
        string event_id
        bool success
    }

    QueryRequest {
        string query
        map<string, string> parameters
        int32 limit
        int32 offset
    }

    MetricSummary {
        string name
        double current_value
        double previous_value
        double change_percent
        Trend trend
        AggregationType aggregation
    }

    MetricSummary ||--|| Trend : has
    MetricSummary ||--|| AggregationType : has
    DimensionValue {
        string value
        int64 count
        double percentage
        map<string, double> metrics
    }

    GetMetricsResponse {
        MetricSeries metrics
    }

    GetMetricsResponse ||--o{ MetricSeries : has
    StreamMetricsRequest {
        string metric_names
        Duration interval
    }

    MetricUpdate {
        Metric metric
        Timestamp timestamp
    }

    MetricUpdate ||--|| Metric : has
    CreateDashboardResponse {
        Dashboard dashboard
    }

    CreateDashboardResponse ||--|| Dashboard : has
    TimeRange {
        Timestamp start
        Timestamp end
        TimeGranularity granularity
    }

    TimeRange ||--|| TimeGranularity : has
    Resolution {
        int32 width
        int32 height
    }

    Size {
        int32 width
        int32 height
    }

    SessionContext {
        string session_id
        Timestamp started_at
        Duration duration
        int32 page_views
        int32 event_count
    }

    BatchTrackResponse {
        int32 events_tracked
        int32 failed_count
    }

    GetDashboardRequest {
        string dashboard_id
    }

    Event {
        string event_id
        EventType event_type
        string event_name
        Timestamp timestamp
        UserContext user
        SessionContext session
        DeviceContext device
        LocationContext location
        map<string, string> properties
        double value
        string currency
        UTMContext utm
        string referrer
        map<string, string> dimensions
    }

    Event ||--|| EventType : has
    Event ||--|| UserContext : has
    Event ||--|| SessionContext : has
    Event ||--|| DeviceContext : has
    Event ||--|| LocationContext : has
    Event ||--|| UTMContext : has
    Report {
        string report_id
        string name
        ReportType type
        TimeRange time_range
        MetricSummary metrics
        Dimension dimensions
        Chart charts
        DataTable tables
        Timestamp generated_at
    }

    Report ||--|| ReportType : has
    Report ||--|| TimeRange : has
    Report ||--o{ MetricSummary : has
    Report ||--o{ Dimension : has
    Report ||--o{ Chart : has
    Report ||--o{ DataTable : has
    Column {
        string name
        DataType type
        string format
    }

    Column ||--|| DataType : has
    GetReportRequest {
        ReportType type
        TimeRange time_range
        map<string, string> filters
        string metrics
        string dimensions
    }

    GetReportRequest ||--|| ReportType : has
    GetReportRequest ||--|| TimeRange : has
    UserContext {
        string user_id
        string anonymous_id
        string email
        map<string, string> traits
    }

    GetMetricsRequest {
        string metric_names
        TimeRange time_range
        map<string, string> filters
        string group_by
    }

    GetMetricsRequest ||--|| TimeRange : has
    MetricSeries {
        string name
        DataPoint points
        map<string, string> metadata
    }

    MetricSeries ||--o{ DataPoint : has
    Cell {
        string value
        string formatted_value
    }

    TrackEventRequest {
        Event event
    }

    TrackEventRequest ||--|| Event : has
    Chart {
        string chart_id
        ChartType type
        string title
        MetricSeries series
        map<string, string> config
    }

    Chart ||--|| ChartType : has
    Chart ||--o{ MetricSeries : has
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

    pb "analytics.v1"
)

func main() {
    // Connect to the service
    conn, err := grpc.Dial("localhost:50051",
        grpc.WithTransportCredentials(insecure.NewCredentials()))
    if err != nil {
        log.Fatalf("Failed to connect: %v", err)
    }
    defer conn.Close()

    client := pb.NewAnalyticsServiceClient(conn)

    // Example RPC call
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    req := &pb.TrackEventRequest{
        // Fill in request fields
    }

    resp, err := client.TrackEvent(ctx, req)
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
import { ProtoGrpcType } from './analytics/analytics';
import { AnalyticsServiceClient } from './analytics.v1/AnalyticsService';

// Load proto file
const packageDefinition = protoLoader.loadSync(
    'analytics/analytics.proto',
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
const client: AnalyticsServiceClient = new proto.analytics.v1.AnalyticsService(
    'localhost:50051',
    grpc.credentials.createInsecure()
);

// Example RPC call
const request = {
    // Fill in request fields
};

client.TrackEvent(request, (error: grpc.ServiceError | null, response?: any) => {
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
