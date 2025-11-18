Title: Add optional local durable queue (PoC) to improve offline reliability

Summary
-------
This PR adds a minimal proof‑of‑concept local durable queue (`pkg/localqueue`) intended to reduce message loss when edge/fog device services experience intermittent connectivity. The queue is opt‑in and uses `go.etcd.io/bbolt` as a pure‑Go storage backend to keep cross‑platform builds simple (Windows friendly).

What this PR includes
---------------------
- `pkg/localqueue/queue.go` — bbolt backed queue with APIs: `Enqueue`, `DequeuePending`, `MarkSent`, `Count`.
- `pkg/localqueue/queue_test.go` — unit tests that simulate temporary send failures and verify enqueue→retry→mark‑sent behavior.
- Integration into Device SDK send path (internal/common/utils.go): on MessageBus publish failure, events are enqueued when `LocalQueue.Enabled=true`.
- Background worker that retries queued events and marks them as sent on success.

Related issue
-------------
Addresses discussion: https://github.com/edgexfoundry/edgex-go/issues/5310

How to test locally
--------------------
1. From repo root:
```powershell
go get go.etcd.io/bbolt@v1.3.10
go mod tidy
go test ./pkg/localqueue -v
go test ./... -v
```
2. To exercise the offline behavior manually:
   - Enable local queue in configuration (set `LocalQueue.Enabled=true` and optionally `LocalQueue.DBPath`).
   - Start a device service and block MessageBus connectivity (or configure MessageBus to be unavailable).
   - Generate events on the device service; they should be enqueued locally.
   - Restore connectivity; background worker should retry and publish enqueued events.

Design notes and future work
---------------------------
- This is a PoC and is intentionally opt‑in.
- Suggested follow‑ups: idempotency tokens for safe retries, TTL/compaction policy, integration tests simulating partitions, configuration defaults and documentation updates.

Request for maintainers
-----------------------
- Please review the PoC: is `bbolt` acceptable or prefer SQLite? Should this be enabled by default?
- Guidance on preferred config keys and where to place the queue wiring in the SDK.

Files changed summary
---------------------
- Added `pkg/localqueue/*` (implementation + tests)
- Modified `internal/config/config.go` to add `LocalQueue` config
- Modified `internal/common/utils.go` to enqueue on publish failure and add retry worker

Thanks — I look forward to feedback and will iterate quickly on requested changes.
