# Architecture: Realtime Direct Messaging & WebSocket Gateway

## 1. Gateway & Connection Lifecycle
The Realtime Gateway exposes a WebSocket endpoint at `/api/v1/ws` protected via session authentication (JWT or session token passed in query parameter `?token=...` or `Sec-WebSocket-Protocol`).

### Connection State Machine
```mermaid
stateDiagram-v2
    [*] --> Connecting
    Connecting --> Connected : Auth Success & Handshake OK
    Connecting --> Disconnected : Auth Failure (401/403)
    Connected --> Heartbeat : Ping every 30s
    Heartbeat --> Connected : Pong received
    Connected --> Reconnecting : Socket Error / Network Drop
    Reconnecting --> Connected : Exponential Backoff Retry Success
    Reconnecting --> Disconnected : Max Retries Exceeded / Explicit Close
    Disconnected --> [*]
```

## 2. Event Contracts

### Client-to-Server Events
- `conversation.subscribe`: `{ conversation_id: string }`
- `conversation.unsubscribe`: `{ conversation_id: string }`
- `typing.start`: `{ conversation_id: string }`
- `typing.stop`: `{ conversation_id: string }`
- `message.send`: `{ conversation_id: string, client_message_id: string, content: string, media_id?: string }`
- `message.read`: `{ conversation_id: string, message_id: string }`
- `ping`: `{}`

### Server-to-Client Events
- `message.created`: Full message payload with author details
- `message.status.updated`: Idempotency acknowledgement `{ client_message_id, id, status: "sent" }`
- `message.read`: `{ conversation_id, reader_id, message_id, read_at }`
- `typing.updated`: `{ conversation_id, user_id, is_typing: boolean }`
- `notification.created`: Full notification payload
- `notification.count.updated`: `{ unread_count: number }`
- `pong`: `{}`

## 3. Data Model for Messaging
```mermaid
erDiagram
    conversations ||--o{ conversation_members : "has"
    conversations ||--o{ messages : "contains"
    messages ||--o{ message_attachments : "has"
    messages ||--o{ message_reactions : "receives"
    conversation_members ||--o{ conversation_read_states : "tracks"

    conversations {
        uuid id PK
        boolean is_group
        varchar name
        timestamptz created_at
        timestamptz updated_at
    }

    conversation_members {
        uuid conversation_id PK,FK
        uuid user_id PK,FK
        timestamptz joined_at
        timestamptz last_read_at
        uuid last_read_message_id
    }

    messages {
        uuid id PK
        uuid conversation_id FK
        uuid sender_id FK
        varchar client_message_id
        text content
        varchar message_type
        uuid reply_to_id
        timestamptz created_at
        timestamptz edited_at
        timestamptz deleted_at
    }
```
