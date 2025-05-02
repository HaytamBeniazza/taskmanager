package main

import (
	"database/sql"
	"errors"
	"strings"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

// SQLiteStorage implements Storage using SQLite database
type SQLiteStorage struct {
	db *sql.DB
}

// NewSQLiteStorage creates a new SQLite storage
func NewSQLiteStorage(dbPath string) (*SQLiteStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteStorage{db: db}
	if err := storage.initialize(); err != nil {
		return nil, err
	}

	return storage, nil
}

// initialize creates the necessary tables if they don't exist
func (s *SQLiteStorage) initialize() error {
	// Create tasks table
	_, err := s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tasks (
			id INTEGER PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT,
			completed BOOLEAN,
			created_at TIMESTAMP,
			due_date TIMESTAMP,
			category TEXT,
			priority TEXT
		)
	`)
	if err != nil {
		return err
	}

	// Create tags table
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS tags (
			task_id INTEGER,
			tag TEXT,
			PRIMARY KEY (task_id, tag),
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	// Create notes table
	_, err = s.db.Exec(`
		CREATE TABLE IF NOT EXISTS notes (
			id INTEGER PRIMARY KEY,
			task_id INTEGER,
			content TEXT,
			created_at TIMESTAMP,
			updated_at TIMESTAMP,
			FOREIGN KEY (task_id) REFERENCES tasks(id) ON DELETE CASCADE
		)
	`)
	if err != nil {
		return err
	}

	return nil
}

// Add saves a new task in the database
func (s *SQLiteStorage) Add(task *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dueDate *time.Time
	if !task.DueDate.IsZero() {
		dueDate = &task.DueDate
	}

	// Insert task
	result, err := tx.Exec(
		`INSERT INTO tasks (title, description, completed, created_at, due_date, category, priority)
		VALUES (?, ?, ?, ?, ?, ?, ?)`,
		task.Title, task.Description, task.Completed, task.CreatedAt,
		dueDate, task.Category, string(task.Priority),
	)
	if err != nil {
		return err
	}

	// Get the auto-generated ID
	id, err := result.LastInsertId()
	if err != nil {
		return err
	}
	task.ID = int(id)

	// Insert tags
	for _, tag := range task.Tags {
		_, err = tx.Exec(
			`INSERT INTO tags (task_id, tag) VALUES (?, ?)`,
			task.ID, tag,
		)
		if err != nil {
			return err
		}
	}

	// Insert notes
	for _, note := range task.Notes {
		_, err = tx.Exec(
			`INSERT INTO notes (task_id, content, created_at, updated_at)
			VALUES (?, ?, ?, ?)`,
			task.ID, note.Content, note.CreatedAt, note.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Get retrieves a task by ID
func (s *SQLiteStorage) Get(id int) (*Task, error) {
	// Query task
	var task Task
	var dueDate sql.NullTime
	var priority string

	err := s.db.QueryRow(
		`SELECT id, title, description, completed, created_at, due_date, category, priority
		FROM tasks WHERE id = ?`, id,
	).Scan(
		&task.ID, &task.Title, &task.Description, &task.Completed,
		&task.CreatedAt, &dueDate, &task.Category, &priority,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	if dueDate.Valid {
		task.DueDate = dueDate.Time
	}
	task.Priority = Priority(priority)

	// Query tags
	rows, err := s.db.Query("SELECT tag FROM tags WHERE task_id = ?", id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		if err := rows.Scan(&tag); err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	task.Tags = tags

	// Query notes
	noteRows, err := s.db.Query(
		"SELECT id, content, created_at, updated_at FROM notes WHERE task_id = ?", id,
	)
	if err != nil {
		return nil, err
	}
	defer noteRows.Close()

	var notes []Note
	for noteRows.Next() {
		var note Note
		if err := noteRows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
			return nil, err
		}
		notes = append(notes, note)
	}
	task.Notes = notes

	return &task, nil
}

// GetAll returns all tasks
func (s *SQLiteStorage) GetAll() ([]*Task, error) {
	// Query all tasks
	rows, err := s.db.Query(
		`SELECT id, title, description, completed, created_at, due_date, category, priority
		FROM tasks`,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var dueDate sql.NullTime
		var priority string

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Completed,
			&task.CreatedAt, &dueDate, &task.Category, &priority,
		)
		if err != nil {
			return nil, err
		}

		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		task.Priority = Priority(priority)
		task.Tags = []string{}
		task.Notes = []Note{}

		tasks = append(tasks, &task)
	}

	// For each task, fetch its tags and notes
	for _, task := range tasks {
		// Fetch tags
		tagRows, err := s.db.Query("SELECT tag FROM tags WHERE task_id = ?", task.ID)
		if err != nil {
			return nil, err
		}

		var tags []string
		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				return nil, err
			}
			tags = append(tags, tag)
		}
		tagRows.Close()
		task.Tags = tags

		// Fetch notes
		noteRows, err := s.db.Query(
			"SELECT id, content, created_at, updated_at FROM notes WHERE task_id = ?", task.ID,
		)
		if err != nil {
			return nil, err
		}

		var notes []Note
		for noteRows.Next() {
			var note Note
			if err := noteRows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
				noteRows.Close()
				return nil, err
			}
			notes = append(notes, note)
		}
		noteRows.Close()
		task.Notes = notes
	}

	return tasks, nil
}

// Update modifies an existing task
func (s *SQLiteStorage) Update(task *Task) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	var dueDate *time.Time
	if !task.DueDate.IsZero() {
		dueDate = &task.DueDate
	}

	// Update task
	_, err = tx.Exec(
		`UPDATE tasks SET
			title = ?, description = ?, completed = ?, due_date = ?,
			category = ?, priority = ?
		WHERE id = ?`,
		task.Title, task.Description, task.Completed, dueDate,
		task.Category, string(task.Priority), task.ID,
	)
	if err != nil {
		return err
	}

	// Delete existing tags and notes
	_, err = tx.Exec("DELETE FROM tags WHERE task_id = ?", task.ID)
	if err != nil {
		return err
	}

	_, err = tx.Exec("DELETE FROM notes WHERE task_id = ?", task.ID)
	if err != nil {
		return err
	}

	// Insert tags
	for _, tag := range task.Tags {
		_, err = tx.Exec(
			`INSERT INTO tags (task_id, tag) VALUES (?, ?)`,
			task.ID, tag,
		)
		if err != nil {
			return err
		}
	}

	// Insert notes
	for _, note := range task.Notes {
		_, err = tx.Exec(
			`INSERT INTO notes (task_id, content, created_at, updated_at)
			VALUES (?, ?, ?, ?)`,
			task.ID, note.Content, note.CreatedAt, note.UpdatedAt,
		)
		if err != nil {
			return err
		}
	}

	return tx.Commit()
}

// Delete removes a task by ID
func (s *SQLiteStorage) Delete(id int) error {
	_, err := s.db.Exec("DELETE FROM tasks WHERE id = ?", id)
	return err
}

// Search finds tasks based on a query string
func (s *SQLiteStorage) Search(query string) ([]*Task, error) {
	// Query tasks that match the search query
	rows, err := s.db.Query(
		`SELECT id, title, description, completed, created_at, due_date, category, priority
		FROM tasks
		WHERE title LIKE ? OR description LIKE ?`,
		"%"+query+"%", "%"+query+"%",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []*Task
	for rows.Next() {
		var task Task
		var dueDate sql.NullTime
		var priority string

		err := rows.Scan(
			&task.ID, &task.Title, &task.Description, &task.Completed,
			&task.CreatedAt, &dueDate, &task.Category, &priority,
		)
		if err != nil {
			return nil, err
		}

		if dueDate.Valid {
			task.DueDate = dueDate.Time
		}
		task.Priority = Priority(priority)
		task.Tags = []string{}
		task.Notes = []Note{}

		tasks = append(tasks, &task)
	}

	// For each task, fetch its tags and notes
	for _, task := range tasks {
		// Fetch tags
		tagRows, err := s.db.Query("SELECT tag FROM tags WHERE task_id = ?", task.ID)
		if err != nil {
			return nil, err
		}

		var tags []string
		for tagRows.Next() {
			var tag string
			if err := tagRows.Scan(&tag); err != nil {
				tagRows.Close()
				return nil, err
			}
			tags = append(tags, tag)
		}
		tagRows.Close()
		task.Tags = tags

		// Fetch notes
		noteRows, err := s.db.Query(
			"SELECT id, content, created_at, updated_at FROM notes WHERE task_id = ?", task.ID,
		)
		if err != nil {
			return nil, err
		}

		var notes []Note
		for noteRows.Next() {
			var note Note
			if err := noteRows.Scan(&note.ID, &note.Content, &note.CreatedAt, &note.UpdatedAt); err != nil {
				noteRows.Close()
				return nil, err
			}
			notes = append(notes, note)
		}
		noteRows.Close()
		task.Notes = notes
	}

	return tasks, nil
}

// GetStats returns statistics about tasks
func (s *SQLiteStorage) GetStats() (map[string]interface{}, error) {
	stats := map[string]interface{}{
		"total":      0,
		"completed":  0,
		"pending":    0,
		"overdue":    0,
		"byCategory": make(map[string]int),
		"byPriority": map[string]int{
			"low":    0,
			"medium": 0,
			"high":   0,
		},
	}

	// Count total tasks
	var total int
	err := s.db.QueryRow("SELECT COUNT(*) FROM tasks").Scan(&total)
	if err != nil {
		return nil, err
	}
	stats["total"] = total

	// Count completed tasks
	var completed int
	err = s.db.QueryRow("SELECT COUNT(*) FROM tasks WHERE completed = 1").Scan(&completed)
	if err != nil {
		return nil, err
	}
	stats["completed"] = completed
	stats["pending"] = total - completed

	// Count overdue tasks
	var overdue int
	err = s.db.QueryRow(
		"SELECT COUNT(*) FROM tasks WHERE completed = 0 AND due_date < ? AND due_date IS NOT NULL",
		time.Now(),
	).Scan(&overdue)
	if err != nil {
		return nil, err
	}
	stats["overdue"] = overdue

	// Count tasks by category
	categoryRows, err := s.db.Query(
		"SELECT category, COUNT(*) FROM tasks WHERE category IS NOT NULL GROUP BY category",
	)
	if err != nil {
		return nil, err
	}
	defer categoryRows.Close()

	categoryStats := make(map[string]int)
	for categoryRows.Next() {
		var category string
		var count int
		if err := categoryRows.Scan(&category, &count); err != nil {
			return nil, err
		}
		categoryStats[category] = count
	}
	stats["byCategory"] = categoryStats

	// Count tasks by priority
	priorityRows, err := s.db.Query(
		"SELECT priority, COUNT(*) FROM tasks GROUP BY priority",
	)
	if err != nil {
		return nil, err
	}
	defer priorityRows.Close()

	priorityStats := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}
	for priorityRows.Next() {
		var priority string
		var count int
		if err := priorityRows.Scan(&priority, &count); err != nil {
			return nil, err
		}
		priorityStats[strings.ToLower(priority)] = count
	}
	stats["byPriority"] = priorityStats

	return stats, nil
}

// Close closes the database connection
func (s *SQLiteStorage) Close() error {
	return s.db.Close()
}
