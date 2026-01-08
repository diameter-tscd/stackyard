# Request/Response Flow Architecture

```mermaid
flowchart TB
    Client[Client Request]
    Handler[Handler Function]
    Bind[request.Bind]
    Validate[request.Validate]
    Logic[Business Logic]
    Success[response.Success]
    Error[response.Error]
    Response[JSON Response]

    Client -->|HTTP Request| Handler
    Handler --> Bind
    Bind -->|Parse JSON| Validate
    Validate -->|Invalid| Error
    Validate -->|Valid| Logic
    Logic -->|Success| Success
    Logic -->|Error| Error
    Success --> Response
    Error --> Response
    Response -->|JSON| Client

    style Client fill:#e1f5ff
    style Response fill:#e1f5ff
    style Success fill:#d4edda
    style Error fill:#f8d7da
    style Logic fill:#fff3cd
```

## Response Structure

```mermaid
classDiagram
    class Response {
        +bool success
        +string message
        +interface{} data
        +ErrorDetail error
        +Meta meta
        +int64 timestamp
    }

    class ErrorDetail {
        +string code
        +string message
        +map details
    }

    class Meta {
        +int page
        +int per_page
        +int64 total
        +int total_pages
        +map extra
    }

    Response --> ErrorDetail
    Response --> Meta
```

## Request Validation Flow

```mermaid
sequenceDiagram
    participant C as Client
    participant H as Handler
    participant B as request.Bind
    participant V as Validator
    participant R as response

    C->>H: POST /api/v1/users
    H->>B: Bind(context, &req)
    B->>B: Parse JSON
    B->>V: Validate(req)
    
    alt Validation Failed
        V-->>B: ValidationError
        B-->>H: return error
        H->>R: ValidationError(...)
        R-->>C: 422 with error details
    else Validation Success
        V-->>B: nil
        B-->>H: nil
        H->>H: Process request
        H->>R: Success(data)
        R-->>C: 200 with data
    end
```

## Async Infrastructure Flow

```mermaid
graph TD
    A[HTTP Request] --> B[Handler]
    B --> C[Async Operation]
    C --> D[Worker Pool]
    D --> E[Infrastructure]
    E --> F[Result Channel]
    F --> G[Response]

    style A fill:#e1f5ff
    style C fill:#fff3cd
    style D fill:#ffeaa7
    style E fill:#fdcb6e
    style G fill:#55a3ff
```

## Service Registration Architecture

```mermaid
graph TD
    A[ServiceRegistrar] --> B[Reflection Analysis]
    B --> C{Dependencies?}
    C -->|No| D[Immediate Registration]
    C -->|Yes| E[Wait for Infrastructure]
    E --> F[Infrastructure Ready]
    F --> G[Register Service]
    G --> H[Boot Routes]

    style A fill:#74b9ff
    style B fill:#0984e3
    style D fill:#00b894
    style G fill:#00b894
```

## Package Organization

```mermaid
graph LR
    A[pkg/request] -->|Validates| F[Handler]
    B[pkg/response] -->|Formats| F
    F -->|Uses| G[Service Logic]
    G -->|Uses| H[Async Infrastructure]
    H -->|Worker Pools| I[Infrastructure Managers]
    G -->|Returns| F

    style A fill:#ffeb9c
    style B fill:#9cf09c
    style H fill:#ffeaa7
    style I fill:#fdcb6e
    style F fill:#9cccff
    style G fill:#ff9c9c

## Complete CRUD Example Flow

```mermaid
graph TD
    subgraph "GET /api/v1/users (List)"
        A1[Parse Pagination] --> A2[Fetch Data]
        A2 --> A3[Calculate Meta]
        A3 --> A4[SuccessWithMeta]
    end

    subgraph "GET /api/v1/users/:id (Detail)"
        B1[Get ID from URL] --> B2[Find in DB]
        B2 -->|Found| B3[Success]
        B2 -->|Not Found| B4[NotFound]
    end

    subgraph "POST /api/v1/users (Create)"
        C1[Bind & Validate] --> C2{Valid?}
        C2 -->|No| C3[ValidationError]
        C2 -->|Yes| C4[Create in DB]
        C4 --> C5[Created 201]
    end

    subgraph "PUT /api/v1/users/:id (Update)"
        D1[Get ID + Bind] --> D2{Valid?}
        D2 -->|No| D3[ValidationError]
        D2 -->|Yes| D4[Update in DB]
        D4 --> D5[Success]
    end

    subgraph "DELETE /api/v1/users/:id (Delete)"
        E1[Get ID] --> E2[Delete from DB]
        E2 --> E3[NoContent 204]
    end

    style A4 fill:#d4edda
    style B3 fill:#d4edda
    style B4 fill:#f8d7da
    style C3 fill:#f8d7da
    style C5 fill:#d4edda
    style D3 fill:#f8d7da
    style D5 fill:#d4edda
    style E3 fill:#d4edda
```

## Response Helper Functions Map

```mermaid
mindmap
  root((Response Helpers))
    Success Responses
      Success 200
      SuccessWithMeta 200
      Created 201
      NoContent 204
    Client Errors
      BadRequest 400
      Unauthorized 401
      Forbidden 403
      NotFound 404
      Conflict 409
      ValidationError 422
    Server Errors
      InternalServerError 500
      ServiceUnavailable 503
    Custom
      Error custom code
```

## Validation Tags Reference

```mermaid
graph TB
    subgraph "String Validators"
        S1[required]
        S2[email]
        S3[min max]
        S4[len]
    end

    subgraph "Number Validators"
        N1[gte lte]
        N2[gt lt]
        N3[min max]
    end

    subgraph "Custom Validators"
        C1[phone]
        C2[username]
    end

    subgraph "Choice Validators"
        O1[oneof]
    end

    style S1 fill:#ffcccc
    style N1 fill:#ccffcc
    style C1 fill:#ccccff
    style O1 fill:#ffffcc
```
