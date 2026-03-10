# Distributed Task Scheduler

A distributed task scheduler built from scratch in Go, demonstrating core distributed systems concepts including concurrent worker management, heartbeat-based failure detection, and automatic task reassignment.

## Overview

The system consists of two components:

- **Coordinator** — accepts tasks via HTTP, manages a task queue, tracks worker health via heartbeats, and reassigns tasks when workers die
- **Workers** — poll tasks from a shared channel, execute them concurrently, and send periodic heartbeats to signal liveness

## Architecture

```
HTTP Client
     |
     | POST /task
     v
 Coordinator
  ├── Task Queue (buffered channel)
  ├── Heartbeat Monitor
  ├── Task Assignment Tracker
  └── Dead Worker Detector
       |
   ┌───┴────────────────┐
   v                    v
Worker 0            Worker 1 ...
  ├── Task goroutine    ├── Task goroutine
  └── Heartbeat ticker  └── Heartbeat ticker
```

## Features

- HTTP API for task submission
- Concurrent worker pool with configurable size
- Heartbeat-based liveness detection (5s interval, 10s timeout)
- Automatic task reassignment on worker failure
- Coordinator shuts down if fewer than half the workers remain alive
- Read/write mutex protecting shared state across goroutines
- Graceful worker shutdown via stop channel and `sync.Once`

## Getting Started

**Requirements:** Go 1.21+

```bash
git clone https://github.com/MehmedMulalic/distributed-task-scheduler
cd distributed-task-scheduler
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

### Health check
```bash
curl http://localhost:5000/healthy
```

## How It Works

### Task Flow

1. Client submits a task via `POST /task`
2. Coordinator places it in a buffered channel (queue)
3. An available worker picks it up and sends a `WorkerAssigned` event
4. Coordinator records which worker owns which task
5. Worker completes the task and sends a `WorkerResult` event
6. Coordinator removes the task from the assignment map

### Failure Detection

Each worker runs two goroutines — one for task processing and one for sending heartbeats every 5 seconds. The coordinator runs a background ticker that checks all worker heartbeat timestamps every 5 seconds. If a worker hasn't been heard from in over 10 seconds, it is considered dead.

On worker death:
- The worker is removed from the heartbeat map
- Its assigned task (if any) is put back into the task queue
- If fewer than half the original workers remain alive, the coordinator logs a fatal error and exits

### Concurrency Design

All shared state (`wHeartbeats`, `wTaskAssigned`) is protected by a `sync.RWMutex`. Multiple goroutines can read simultaneously, but writes acquire an exclusive lock. Workers use directional channels (`chan<-`, `<-chan`) to enforce communication boundaries at compile time.

## Project Structure

```
.
├── main.go         # Entry point, wires channels and starts components
├── coordinator.go  # Coordinator logic, HTTP handlers, failure detection
├── worker.go       # Worker logic, task execution, heartbeat sending
└── utils.go        # Shared types (Task, WorkerResult, WorkerHeartbeat, etc.)
```

## Concepts Demonstrated

- Goroutines and channel-based concurrency
- Buffered vs unbuffered channels
- Directional channel types for enforcing communication flow
- `sync.RWMutex` for shared state protection
- `sync.Once` for safe one-time operations (graceful shutdown)
- Heartbeat pattern for distributed liveness detection
- Task reassignment on node failure
