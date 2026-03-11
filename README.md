# Distributed Task Scheduler

A distributed task scheduler built from scratch in Go, demonstrating core distributed systems concepts: concurrent worker management, heartbeat-based failure detection, automatic task reassignment, and SQLite-backed persistence.

## Overview

The system has two components:

- **Coordinator** — accepts tasks via HTTP, persists them to SQLite, feeds pending tasks to workers, tracks worker health via heartbeats, and reassigns tasks when workers die
- **Workers** — receive tasks from a shared channel, execute them concurrently, and send periodic heartbeats to signal liveness

## Architecture

```
HTTP Client
     |
     | POST /task
     v
 Coordinator
  ├── SQLite Database (persistence)
  ├── Feeder goroutine (pending → channel)
  ├── Heartbeat Monitor
  ├── Task Assignment Tracker
  └── Dead Worker Detector
           |
   ┌───────┴──────────────┐
   v                      v
Worker 0              Worker 1 ...
  ├── Task goroutine      ├── Task goroutine
  └── Heartbeat ticker    └── Heartbeat ticker
```

## Features

- HTTP API for task submission and status monitoring
- SQLite persistence — tasks survive coordinator restarts
- Concurrent worker pool with configurable size
- Feeder goroutine bridges database queue to worker channel
- Atomic task state transitions via SQL transactions
- Heartbeat-based liveness detection
- Automatic task reassignment on worker failure
- Coordinator shuts down if fewer than half the workers remain alive
- `sync.RWMutex` protecting all shared in-memory state
- Graceful worker shutdown via stop channel and `sync.Once`

## Getting Started

**Requirements:** Go 1.25+

```bash
git clone https://github.com/MehmedMulalic/Task-Scheduler
cd Task-Scheduler
go run .
```

The coordinator starts on port `5000`.

## API

### Submit a task
```bash
curl -X POST http://localhost:5000/task \
  -H "Content-Type: application/json" \
  -d '{"message": "your task here"}'
```

### View all tasks and their statuses
```bash
curl http://localhost:5000/tasks
```

Example response:
```json
[
  {"task_id": 1, "task_message": "your task here", "status": "completed"},
  {"task_id": 2, "task_message": "another task", "status": "in_progress"}
]
```

### Health check
```bash
curl http://localhost:5000/healthy
```

## Task Lifecycle

```
pending → in_progress → completed
              ↓
          (worker dies)
              ↓
           pending  ← reassigned, picked up by another worker
```

Tasks are stored in SQLite with one of three statuses. The transition from `pending` to `in_progress` happens atomically inside a SQL transaction, preventing two workers from claiming the same task simultaneously.

## How It Works

### Persistence

When a task is submitted via `POST /task`, the coordinator immediately writes it to SQLite as `pending` and returns `201`. No blocking, no channel pressure.

A feeder goroutine polls the database every second for pending tasks and pushes them into the worker channel, updating their status to `in_progress` atomically via a transaction.

### Failure Detection

Each worker runs two goroutines — one for task processing, one for sending heartbeats every 5 seconds. The coordinator checks all heartbeat timestamps every 2 seconds. If a worker hasn't been heard from in over 10 seconds, it is considered dead.

On worker death:
- The worker is removed from the heartbeat map
- Its assigned task status is reset to `pending` in the database
- The feeder picks it up again within the next poll cycle
- If fewer than half the original workers remain alive, the coordinator exits

### Concurrency Design

Shared in-memory state (`wHeartbeats`, `wTaskAssigned`) is protected by `sync.RWMutex` — multiple goroutines can read simultaneously, writes acquire an exclusive lock. Workers use directional channels (`chan<-`, `<-chan`) to enforce communication flow at compile time. `sync.Once` ensures workers can only be stopped once, preventing double-close panics.

## Project Structure

```
.
├── main.go              # Entry point — wires DB, channels, coordinator, workers
├── coordinator.go       # HTTP handlers, worker monitoring, failure detection
├── worker.go            # Task execution, heartbeat sending, graceful shutdown
├── task_repository.go   # All database logic — inserts, queries, status updates
└── utils.go             # Shared types: Task, WorkerResult, WorkerHeartbeat, etc.
```

## Concepts Demonstrated

- Goroutines and channel-based concurrency
- Buffered vs unbuffered channels
- Directional channel types (`chan<-`, `<-chan`)
- `sync.RWMutex` for shared state protection
- `sync.Once` for safe one-time shutdown
- SQL transactions for atomic state transitions
- Heartbeat pattern for distributed liveness detection
- Repository pattern for database abstraction
- Task reassignment on node failure
