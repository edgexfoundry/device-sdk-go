# Local Persistent Queue

This package provides a local durable queue implementation for the EdgeX Device SDK. It enables device services to persist outbound messages locally when network connectivity is unavailable, ensuring reliable message delivery for fog and edge nodes with intermittent connectivity.

## Overview

The local queue addresses the problem of message loss when device services attempt to deliver events/commands to Core services during network outages. Messages are persisted locally using bbolt (a pure-Go key-value store) and automatically retried when connectivity is restored.

## Features

- **Persistent Storage**: Messages are stored in a local bbolt database, surviving device restarts
- **TTL Support**: Configurable time-to-live for queue items to prevent unbounded growth
- **Automatic Cleanup**: Removes expired items and optionally removes old sent items
- **Thread-Safe**: Supports concurrent enqueue/dequeue operations
- **Cross-Platform**: Pure Go implementation, works on Windows, Linux, and macOS without cgo
- **Configurable Limits**: Maximum queue size and cleanup policies

## API

### Creating a Queue

```go
import "github.com/edgexfoundry/device-sdk-go/v4/pkg/localqueue"

// Create a queue with max 10000 items and 7-day TTL (defaults)
queue, err := localqueue.NewLocalQueue("./localqueue.db", 10000, 7*24*time.Hour)
if err != nil {
    log.Fatal(err)
}
defer queue.Close()
```

### Enqueueing Messages

```go
payload := []byte("your message data")
id, err := queue.Enqueue(payload)
if err != nil {
    if err == localqueue.ErrQueueFull {
        // Handle queue full condition
    }
    log.Fatal(err)
}
```

### Dequeueing Pending Messages

```go
// Get up to 10 pending items
items, err := queue.DequeuePending(10)
if err != nil {
    log.Fatal(err)
}

for _, item := range items {
    // Attempt to send
    if err := sendToCore(item.Payload); err != nil {
        // Will be retried on next dequeue
        continue
    }
    
    // Mark as sent on success
    if err := queue.MarkSent(item.ID); err != nil {
        log.Printf("Failed to mark item as sent: %v", err)
    }
}
```

### Cleanup Operations

```go
// Remove expired items and sent items older than 24 hours
removed, err := queue.Cleanup(24 * time.Hour)
if err != nil {
    log.Fatal(err)
}
log.Printf("Removed %d items", removed)
```

### Querying Queue Status

```go
// Count pending items
pendingCount, err := queue.Count()

// Count all items (pending + sent)
totalCount, err := queue.CountAll()

// Get a specific item
item, err := queue.Get(itemID)
```

## Configuration

The queue supports the following configuration options:

- **MaxItems**: Maximum number of pending items (default: 10000)
- **ItemTTL**: Time-to-live for items (default: 7 days)
- **CleanupInterval**: Interval for periodic cleanup operations (default: 1 hour)

## Error Handling

The package defines the following errors:

- `ErrQueueFull`: Returned when the queue has reached its maximum capacity
- `ErrItemNotFound`: Returned when an item with the given ID is not found
- `ErrInvalidItem`: Returned when an item has invalid data

## Testing

Run the test suite:

```bash
go test ./pkg/localqueue -v
```

The test suite includes:
- Basic enqueue/dequeue operations
- TTL expiration handling
- Cleanup operations
- Concurrent access patterns
- Error conditions
- Edge cases

## Integration with Device SDK

This package is intended to be integrated into the Device SDK's send path. When enabled via configuration:

1. On send failure (temporary network error), messages are enqueued locally
2. A background worker periodically dequeues pending items and retries delivery
3. Successfully sent items are marked as sent
4. Periodic cleanup removes expired and old sent items

## Related Issue

This implementation addresses [EdgeX issue #5310](https://github.com/edgexfoundry/edgex-go/issues/5310): "Add local durable queue for intermittent connectivity"

## License

Apache-2.0 (same as EdgeX Foundry)
