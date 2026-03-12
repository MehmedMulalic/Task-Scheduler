package main_test

import (
	"database/sql"
	main "dts"
	"testing"
)

func setupTest(t *testing.T) *main.TaskRepository {
	t.Helper()

	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	_, err = db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			message TEXT NOT NULL,
			status TEXT NOT NULL DEFAULT 'pending'
		)
	`)
	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		db.Close()
	})

	return main.NewTaskRepository(db)
}

func TestCreateTask(t *testing.T) {
	repo := setupTest(t)

	err := repo.CreateTask(main.Task{Message: "test message"})
	if err != nil {
		t.Fatal(err)
	}

	tasks, err := repo.GetAllTasks()
	if err != nil {
		t.Fatal(err)
	}

	if len(tasks) != 1 {
		t.Errorf("expected 1 task, got %d", len(tasks))
	}

	if tasks[0].Message != "test message" {
		t.Errorf("expected 'test message', got '%s'", tasks[0].Message)
	}

	if tasks[0].Status != main.Pending {
		t.Errorf("expected pending status, got %s", tasks[0].Status)
	}
}

func TestGetPendingTask_ReturnNilWhenEmpty(t *testing.T) {
	repo := setupTest(t)

	task, err := repo.GetPendingTask()
	if err != nil {
		t.Fatal(err)
	}
	if task != nil {
		t.Fatalf("expected nil, got %+v", task)
	}
}

func TestGetPendingTask_TaskInProgress(t *testing.T) {
	repo := setupTest(t)

	repo.CreateTask(main.Task{Message: "test"})

	task, err := repo.GetPendingTask()
	if err != nil {
		t.Fatal(err)
	}
	if task == nil {
		t.Fatal("expected a task, got nil")
	}

	tasks, _ := repo.GetAllTasks()
	if tasks[0].Status != main.InProgress {
		t.Errorf("expected in_progress, got %s", tasks[0].Status)
	}
}

func TestGetCompletedTask(t *testing.T) {
	repo := setupTest(t)

	repo.CreateTask(main.Task{Message: "test"})
	repo.UpdateTaskStatus(1, main.Completed)

	tasks, _ := repo.GetAllTasks()
	if tasks[0].Status != main.Completed {
		t.Errorf("expected completed, got %s", tasks[0].Status)
	}
}

func TestGetTaskById(t *testing.T) {
	repo := setupTest(t)

	repo.CreateTask(main.Task{Message: "test"})

	task, _ := repo.GetTaskById(1)
	if task.Id != 1 {
		t.Errorf("expected 1, got %d", task.Id)
	}
}

func TestWorkerDeath(t *testing.T) {
	repo := setupTest(t)

	repo.CreateTask(main.Task{Message: "test"})

	task, err := repo.GetPendingTask()
	if err != nil {
		t.Fatal(err)
	}
	t.Log(task)

	repo.UpdateTaskStatus(task.Id, main.Pending)

	task, err = repo.GetPendingTask()
	t.Log(task)
	if err != nil {
		t.Fatal(err)
	}
	if task == nil {
		t.Fatal("expected task to be reassigned, got nil")
	}

	if task.Status != main.InProgress {
		t.Errorf("expected in_progress, got %s", task.Status)
	}
}
