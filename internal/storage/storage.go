package storage

import (
	"sort"
	"strings"
	"time"

	"github.com/HaytamBeniazza/taskmanager/internal/models"
)

// FilterOptions defines options for filtering tasks
type FilterOptions struct {
	Completed    *bool           `json:"completed,omitempty"`    // Filter by completion status
	Category     string          `json:"category,omitempty"`     // Filter by category
	Priority     models.Priority `json:"priority,omitempty"`     // Filter by priority
	Tag          string          `json:"tag,omitempty"`          // Filter by tag
	DueBefore    time.Time       `json:"dueBefore,omitempty"`    // Filter by due date before
	DueAfter     time.Time       `json:"dueAfter,omitempty"`     // Filter by due date after
	CreatedAfter time.Time       `json:"createdAfter,omitempty"` // Filter by creation date after
	AssignedTo   string          `json:"assignedTo,omitempty"`   // Filter by assignee
	CreatedBy    string          `json:"createdBy,omitempty"`    // Filter by creator
}

// SortOptions defines options for sorting tasks
type SortOptions struct {
	Field     string `json:"field"`     // Field to sort by (e.g., "dueDate", "priority", "title")
	Direction string `json:"direction"` // Sort direction ("asc" or "desc")
}

// PaginationOptions defines pagination parameters for list queries
type PaginationOptions struct {
	Page    int `json:"page"`
	PerPage int `json:"perPage"`
}

// TaskStorage defines the interface for task storage
type TaskStorage interface {
	// Basic CRUD operations
	Add(task *models.Task) error
	Get(id int) (*models.Task, error)
	GetAll() ([]*models.Task, error)
	Update(task *models.Task) error
	Delete(id int) error

	// Advanced query operations
	Search(query string) ([]*models.Task, error)
	Filter(options FilterOptions) ([]*models.Task, error)
	GetPaginated(page, perPage int, sortOptions SortOptions) ([]*models.Task, int, error)

	// Statistics and metadata
	GetStats() (map[string]interface{}, error)
	GetCategories() ([]string, error)
	GetTags() ([]string, error)
	GetTaskAnalytics() (map[string]interface{}, error)

	// Attachment operations
	SaveAttachment(taskID int, filename string, data []byte, contentType string) (*models.Attachment, error)
	GetAttachment(taskID int, attachmentID int) (*models.Attachment, []byte, error)
	DeleteAttachment(taskID int, attachmentID int) error

	// User-related operations
	GetTasksByUser(username string) ([]*models.Task, error)
	GetSharedTasks(username string) ([]*models.Task, error)

	// Batch operations
	BatchCreate(tasks []*models.Task) error
	BatchUpdate(tasks []*models.Task) error
	BatchDelete(ids []int) error
	BatchComplete(ids []int) error

	// Database management
	Close() error
}

// UserStorage defines the interface for user storage
type UserStorage interface {
	// User operations
	AddUser(user *models.User) error
	GetUser(id int) (*models.User, error)
	GetUserByUsername(username string) (*models.User, error)
	GetUserByEmail(email string) (*models.User, error)
	UpdateUser(user *models.User) error
	DeleteUser(id int) error
	ListUsers() ([]*models.User, error)

	// Authentication
	Authenticate(username, password string) (*models.User, error)
	ResetPassword(userID int, newPassword string) error

	// Close connection
	Close() error
}

// Helper function for filtering tasks
func FilterTasks(tasks []*models.Task, options FilterOptions) []*models.Task {
	filtered := make([]*models.Task, 0)

	for _, task := range tasks {
		// Filter by completion status
		if options.Completed != nil && task.Completed != *options.Completed {
			continue
		}

		// Filter by category
		if options.Category != "" && task.Category != options.Category {
			continue
		}

		// Filter by priority
		if options.Priority != "" && task.Priority != options.Priority {
			continue
		}

		// Filter by tag
		if options.Tag != "" {
			hasTag := false
			for _, tag := range task.Tags {
				if tag == options.Tag {
					hasTag = true
					break
				}
			}
			if !hasTag {
				continue
			}
		}

		// Filter by due date before
		if !options.DueBefore.IsZero() && (task.DueDate.IsZero() || !task.DueDate.Before(options.DueBefore)) {
			continue
		}

		// Filter by due date after
		if !options.DueAfter.IsZero() && (task.DueDate.IsZero() || !task.DueDate.After(options.DueAfter)) {
			continue
		}

		// Filter by created after
		if !options.CreatedAfter.IsZero() && !task.CreatedAt.After(options.CreatedAfter) {
			continue
		}

		// Filter by assignee
		if options.AssignedTo != "" && task.AssignedTo != options.AssignedTo {
			continue
		}

		// Filter by creator
		if options.CreatedBy != "" && task.CreatedBy != options.CreatedBy {
			continue
		}

		filtered = append(filtered, task)
	}

	return filtered
}

// Helper function for sorting tasks
func SortTasks(tasks []*models.Task, options SortOptions) []*models.Task {
	sorted := make([]*models.Task, len(tasks))
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
		priorityMap := map[models.Priority]int{
			models.Low:    1,
			models.Medium: 2,
			models.High:   3,
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

// Helper function for string contains search
func Contains(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}
