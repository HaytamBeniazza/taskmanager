package main

import (
	"sort"
	"strings"
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

// FilterOptions defines options for filtering tasks
type FilterOptions struct {
	Completed   *bool     `json:"completed,omitempty"`   // Filter by completion status
	Category    string    `json:"category,omitempty"`    // Filter by category
	Priority    Priority  `json:"priority,omitempty"`    // Filter by priority
	Tag         string    `json:"tag,omitempty"`         // Filter by tag
	DueBefore   time.Time `json:"dueBefore,omitempty"`   // Filter by due date before
	DueAfter    time.Time `json:"dueAfter,omitempty"`    // Filter by due date after
	CreatedAfter time.Time `json:"createdAfter,omitempty"` // Filter by creation date after
	AssignedTo   string    `json:"assignedTo,omitempty"`   // Filter by assignee
	CreatedBy    string    `json:"createdBy,omitempty"`    // Filter by creator
}

// SortOptions defines options for sorting tasks
type SortOptions struct {
	Field     string `json:"field"`     // Field to sort by (e.g., "dueDate", "priority", "title")
	Direction string `json:"direction"` // Sort direction ("asc" or "desc")
}

// Storage defines the interface for task storage
type Storage interface {
	// Basic CRUD operations
	Add(task *Task) error
	Get(id int) (*Task, error)
	GetAll() ([]*Task, error)
	Update(task *Task) error
	Delete(id int) error
	
	// Advanced query operations
	Search(query string) ([]*Task, error)
	Filter(options FilterOptions) ([]*Task, error)
	Sort(tasks []*Task, options SortOptions) []*Task
	
	// Statistics and metadata
	GetStats() (map[string]interface{}, error)
	GetCategories() ([]string, error)
	GetTags() ([]string, error)
	
	// Attachment operations
	SaveAttachment(taskID int, filename string, data []byte, contentType string) (*Attachment, error)
	GetAttachment(taskID int, attachmentID int) (*Attachment, []byte, error)
	DeleteAttachment(taskID int, attachmentID int) error
	
	// User-related operations
	GetTasksByUser(username string) ([]*Task, error)
	GetSharedTasks(username string) ([]*Task, error)
	
	// Database management
	Close() error
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

// SortTasks sorts a slice of tasks based on the given sort options
func SortTasks(tasks []*Task, options SortOptions) []*Task {
	sorted := make([]*Task, len(tasks))
	copy(sorted, tasks)
	
	switch options.Field {
	case "dueDate":
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				// Handle nil due dates (tasks without due dates come last)
				if sorted[i].DueDate.IsZero() && !sorted[j].DueDate.IsZero() {
					return false
				}
				if !sorted[i].DueDate.IsZero() && sorted[j].DueDate.IsZero() {
					return true
				}
				if sorted[i].DueDate.IsZero() && sorted[j].DueDate.IsZero() {
					return sorted[i].ID < sorted[j].ID
				}
				return sorted[i].DueDate.Before(sorted[j].DueDate)
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				// Handle nil due dates (tasks without due dates come last)
				if sorted[i].DueDate.IsZero() && !sorted[j].DueDate.IsZero() {
					return false
				}
				if !sorted[i].DueDate.IsZero() && sorted[j].DueDate.IsZero() {
					return false
				}
				if sorted[i].DueDate.IsZero() && sorted[j].DueDate.IsZero() {
					return sorted[i].ID > sorted[j].ID
				}
				return sorted[j].DueDate.Before(sorted[i].DueDate)
			})
		}
	case "priority":
		priorityMap := map[Priority]int{
			Low:    1,
			Medium: 2,
			High:   3,
		}
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				return priorityMap[sorted[i].Priority] < priorityMap[sorted[j].Priority]
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return priorityMap[sorted[i].Priority] > priorityMap[sorted[j].Priority]
			})
		}
	case "title":
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Title < sorted[j].Title
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Title > sorted[j].Title
			})
		}
	case "createdAt":
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].CreatedAt.Before(sorted[j].CreatedAt)
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[j].CreatedAt.Before(sorted[i].CreatedAt)
			})
		}
	case "category":
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Category < sorted[j].Category
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].Category > sorted[j].Category
			})
		}
	default:
		// Default sort by ID
		if options.Direction == "asc" {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].ID < sorted[j].ID
			})
		} else {
			sort.Slice(sorted, func(i, j int) bool {
				return sorted[i].ID > sorted[j].ID
			})
		}
	}
	
	return sorted
}

// FilterTasks filters a slice of tasks based on the given filter options
func FilterTasks(tasks []*Task, options FilterOptions) []*Task {
	result := make([]*Task, 0)
	
	for _, task := range tasks {
		// Check all filter conditions
		if options.Completed != nil && task.Completed != *options.Completed {
			continue
		}
		
		if options.Category != "" && task.Category != options.Category {
			continue
		}
		
		if options.Priority != "" && task.Priority != options.Priority {
			continue
		}
		
		if options.Tag != "" {
			found := false
			for _, tag := range task.Tags {
				if tag == options.Tag {
					found = true
					break
				}
			}
			if !found {
				continue
			}
		}
		
		if !options.DueBefore.IsZero() && !task.DueDate.IsZero() && !task.DueDate.Before(options.DueBefore) {
			continue
		}
		
		if !options.DueAfter.IsZero() && !task.DueDate.IsZero() && !task.DueDate.After(options.DueAfter) {
			continue
		}
		
		if !options.CreatedAfter.IsZero() && !task.CreatedAt.After(options.CreatedAfter) {
			continue
		}
		
		if options.AssignedTo != "" && task.AssignedTo != options.AssignedTo {
			continue
		}
		
		if options.CreatedBy != "" && task.CreatedBy != options.CreatedBy {
			continue
		}
		
		// If we get here, the task passes all filters
		result = append(result, task)
	}
	
	return result
}

func contains(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
