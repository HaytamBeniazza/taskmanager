package main

import (
	"sync"
	"time"
)

// TaskStorage defines the interface for task storage
type TaskStorage interface {
	Add(Task) (Task, error)
	Update(Task) error
	Delete(int) error
	Get(int) (Task, error)
	GetAll() []Task
}

// Storage defines the interface for task storage
type Storage interface {
	Add(task *Task) error
	Get(id int) (*Task, error)
	GetAll() ([]*Task, error)
	Update(task *Task) error
	Delete(id int) error
	Search(query string) ([]*Task, error)
	GetStats() (map[string]interface{}, error)
}

// InMemoryStorage implements TaskStorage using in-memory data structures
type InMemoryStorage struct {
	tasks  map[int]*Task
	mu     sync.RWMutex
	nextID int
}

// NewInMemoryStorage creates a new in-memory storage
func NewInMemoryStorage() *InMemoryStorage {
	return &InMemoryStorage{
		tasks:  make(map[int]*Task),
		nextID: 1,
	}
}

// Add saves a new task in memory
func (s *InMemoryStorage) Add(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	task.ID = s.nextID
	s.nextID++
	s.tasks[task.ID] = task
	return nil
}

// Update modifies an existing task
func (s *InMemoryStorage) Update(task *Task) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.tasks[task.ID]; !exists {
		return nil
	}
	s.tasks[task.ID] = task
	return nil
}

// Delete removes a task by ID
func (s *InMemoryStorage) Delete(id int) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.tasks, id)
	return nil
}

// Get retrieves a task by ID
func (s *InMemoryStorage) Get(id int) (*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	task, exists := s.tasks[id]
	if !exists {
		return nil, nil
	}
	return task, nil
}

// GetAll returns all tasks
func (s *InMemoryStorage) GetAll() ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tasks := make([]*Task, 0, len(s.tasks))
	for _, task := range s.tasks {
		tasks = append(tasks, task)
	}
	return tasks, nil
}

func (s *InMemoryStorage) Search(query string) ([]*Task, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	results := make([]*Task, 0)
	for _, task := range s.tasks {
		if contains(task.Title, query) || contains(task.Description, query) {
			results = append(results, task)
		}
	}
	return results, nil
}

func (s *InMemoryStorage) GetStats() (map[string]interface{}, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	stats := map[string]interface{}{
		"total":      len(s.tasks),
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

	for _, task := range s.tasks {
		if task.Completed {
			stats["completed"] = stats["completed"].(int) + 1
		} else {
			stats["pending"] = stats["pending"].(int) + 1
		}

		if !task.Completed && !task.DueDate.IsZero() && task.DueDate.Before(time.Now()) {
			stats["overdue"] = stats["overdue"].(int) + 1
		}

		stats["byCategory"].(map[string]int)[task.Category]++
		stats["byPriority"].(map[string]int)[string(task.Priority)]++
	}

	return stats, nil
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && s[:len(substr)] == substr
}
