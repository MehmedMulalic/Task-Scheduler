package main

import "database/sql"

type TaskRepository struct {
	db *sql.DB
}

func NewTaskRepository(db *sql.DB) *TaskRepository {
	return &TaskRepository{db}
}

func (r *TaskRepository) CreateTask(t Task) error {
	_, err := r.db.Exec("INSERT INTO tasks (message) VALUES (?)", t.Message)
	if err != nil {
		return err
	}
	return nil
}

func (r *TaskRepository) GetTaskById(id int) (*TaskResponse, error) {
	result := r.db.QueryRow("SELECT id, message FROM tasks WHERE id = (?)", id)

	t := &TaskResponse{}
	err := result.Scan(&t.Id, &t.Message)
	if err != nil {
		return nil, err
	}

	return t, nil
}

func (r *TaskRepository) GetAllTasks() ([]*TaskResponse, error) {
	result, err := r.db.Query("SELECT id, message, status FROM tasks")
	if err != nil {
		return nil, err
	}
	defer result.Close()

	tasks := []*TaskResponse{}

	for result.Next() {
		t := &TaskResponse{}

		err := result.Scan(&t.Id, &t.Message, &t.Status)
		if err != nil {
			return nil, err
		}

		tasks = append(tasks, t)
	}

	return tasks, nil
}

func (r *TaskRepository) GetPendingTask() (*TaskResponse, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	result := tx.QueryRow("SELECT id, message, status FROM tasks WHERE status = 'pending' LIMIT 1")

	t := &TaskResponse{}
	err = result.Scan(&t.Id, &t.Message, &t.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	_, err = tx.Exec("UPDATE tasks SET status = 'in_progress' WHERE id = (?)", t.Id)
	if err != nil {
		return nil, err
	}

	tx.Commit()
	return t, nil
}

func (r *TaskRepository) UpdateTaskStatus(id int, status TaskStatus) (int64, error) {
	result, err := r.db.Exec("UPDATE tasks SET status = ? WHERE id = ?", status, id)
	if err != nil {
		return 0, err
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, err
	}

	return rowsAffected, nil
}
