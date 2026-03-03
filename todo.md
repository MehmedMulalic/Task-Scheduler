Stage 1 — Single node, no distribution
Build a coordinator that accepts tasks via HTTP and a single worker that polls for tasks. No failure handling yet. Just get the basic loop working: submit task → worker gets task → worker completes task → coordinator knows it's done.
This alone teaches you Go HTTP, goroutines for handling concurrent requests, and basic queue logic.

Stage 2 — Multiple workers
Run 3 workers simultaneously. Now you have a real concurrency problem — two workers might try to grab the same task. This forces you to think about locking your queue properly. This is where your concurrency knowledge becomes practical.

Stage 3 — Heartbeats and failure detection
Add the heartbeat system. Each worker sends a ping every 5 seconds. The coordinator runs a background goroutine checking timestamps. If a worker goes silent, reassign its task. Simulate a crash by just killing a worker process and watch the coordinator recover.

Stage 4 — Persistence (optional but impressive)
Right now if the coordinator crashes, all tasks are lost. Add simple file or SQLite persistence so tasks survive restarts.
