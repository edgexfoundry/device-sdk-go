Issue: Use existing upstream discussion for intermittent connectivity reliability

Related upstream issue (created by user): https://github.com/edgexfoundry/edgex-go/issues/5310

Summary
-------
This document links the existing upstream discussion about improving reliability for intermittent connectivity and describes the PoC added in this fork.

What I implemented (PoC)
------------------------
- `pkg/localqueue`: a minimal local durable queue backed by `go.etcd.io/bbolt`.
- Integrates opt-in behavior: on MessageBus publish failure, the Device SDK enqueues the event for later retry.
- A background worker retries pending queued events and marks them as sent on success.
- Unit tests included for the `localqueue` package.

How this relates to upstream issue
---------------------------------
The PoC addresses the problem described in https://github.com/edgexfoundry/edgex-go/issues/5310 by adding an opt-in local persistence layer to avoid message loss during short network outages.

What I need from maintainers
----------------------------
- Feedback on storage backend choice (`bbolt` vs `SQLite`).
- Guidance on default behavior (enabled vs disabled by default).
- Advice on preferred integration points and API surface for the SDK (idempotency keys, TTL, compaction policy).

How to test locally
-------------------
1. Ensure Go 1.20+ is installed.
2. From repo root:
```powershell
go get go.etcd.io/bbolt@v1.3.10
go mod tidy
go test ./pkg/localqueue -v
```

Link to this fork/branch
------------------------
https://github.com/lubitelkvokk/device-sdk-go/tree/fix/offline-queue-181125

Next steps
----------
1. Submit PR with PoC and tests (this repo/branch). Reference upstream issue 5310 in PR body.
2. Iterate with maintainers: wiring into SDK behind config flag, add idempotency, add integration tests.
