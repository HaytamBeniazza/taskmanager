package storage

import (
	"sort"
	"strings"
	"time"

	"github.com/HaytamBeniazza/taskmanager/internal/models"
)

// FilterOptions defines options for filtering tasks
type FilterOptions struct {
	Completed     *bool             `json:"completed,omitempty"`     // Filter by completion status
	Category      string            `json:"category,omitempty"`      // Filter by category
	Priority      models.Priority   `json:"priority,omitempty"`      // Filter by priority
	Tag           string            `json:"tag,omitempty"`           // Filter by tag
	DueBefore     time.Time         `json:"dueBefore,omitempty"`     // Filter by due date before
	DueAfter      time.Time         `json:"dueAfter,omitempty"`      // Filter by due date after
	CreatedAfter  time.Time         `json:"createdAfter,omitempty"`  // Filter by creation date after
	AssignedTo    string            `json:"assignedTo,omitempty"`    // Filter by assignee
	CreatedBy     string            `json:"createdBy,omitempty"`     // Filter by creator
	SearchTerm    string            `json:"searchTerm,omitempty"`    // Search term to match against title and description
	CustomFilters map[string]string `json:"customFilters,omitempty"` // Additional custom filters
}

// SortOptions defines options for sorting tasks
type SortOptions struct {
	Field     string `json:"field"`     // Field to sort by
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
	Close() error

	// Search and filtering
	Search(query string) ([]*models.Task, error)
	Filter(options FilterOptions) ([]*models.Task, error)
	GetCategories() ([]string, error)
	GetTags() ([]string, error)

	// Pagination and sorting
	GetPaginated(page, perPage int, sortOptions SortOptions) ([]*models.Task, int, error)

	// Statistics and analytics
	GetStats() (map[string]interface{}, error)
	GetTaskAnalytics() (map[string]interface{}, error)

	// Batch operations
	BatchCreate(tasks []*models.Task) error
	BatchUpdate(tasks []*models.Task) error
	BatchDelete(ids []int) error
	BatchComplete(ids []int) error

	// User-specific operations
	GetTasksByUser(username string) ([]*models.Task, error)
	GetSharedTasks(username string) ([]*models.Task, error)

	// Attachments
	SaveAttachment(taskID int, filename string, data []byte, contentType string) (*models.Attachment, error)
	GetAttachment(taskID int, attachmentID int) (*models.Attachment, []byte, error)
	DeleteAttachment(taskID int, attachmentID int) error
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

// FilterTasks applies filtering options to a slice of tasks
func FilterTasks(tasks []*models.Task, options FilterOptions) []*models.Task {
	var filteredTasks []*models.Task

	for _, task := range tasks {
		// Check if task matches all filter criteria
		matches := true

		// Filter by completion status
		if options.Completed != nil && task.Completed != *options.Completed {
			matches = false
		}

		// Filter by category
		if options.Category != "" && !strings.EqualFold(task.Category, options.Category) {
			matches = false
		}

		// Filter by priority
		if options.Priority != "" && task.Priority != options.Priority {
			matches = false
		}

		// Filter by tag
		if options.Tag != "" {
			tagFound := false
			normalizedTag := strings.ToLower(strings.TrimSpace(options.Tag))
			for _, tag := range task.Tags {
				if strings.ToLower(strings.TrimSpace(tag)) == normalizedTag {
					tagFound = true
					break
				}
			}
			if !tagFound {
				matches = false
			}
		}

		// Filter by due date before
		if !options.DueBefore.IsZero() && (task.DueDate.IsZero() || task.DueDate.After(options.DueBefore)) {
			matches = false
		}

		// Filter by due date after
		if !options.DueAfter.IsZero() && (task.DueDate.IsZero() || task.DueDate.Before(options.DueAfter)) {
			matches = false
		}

		// Filter by creation date after
		if !options.CreatedAfter.IsZero() && task.CreatedAt.Before(options.CreatedAfter) {
			matches = false
		}

		// Filter by assignee
		if options.AssignedTo != "" && task.AssignedTo != options.AssignedTo {
			matches = false
		}

		// Filter by creator
		if options.CreatedBy != "" && task.CreatedBy != options.CreatedBy {
			matches = false
		}

		// Filter by search term
		if options.SearchTerm != "" {
			term := strings.ToLower(options.SearchTerm)
			titleMatch := strings.Contains(strings.ToLower(task.Title), term)
			descMatch := strings.Contains(strings.ToLower(task.Description), term)
			catMatch := strings.Contains(strings.ToLower(task.Category), term)

			if !titleMatch && !descMatch && !catMatch {
				// Also search in tags
				tagMatch := false
				for _, tag := range task.Tags {
					if strings.Contains(strings.ToLower(tag), term) {
						tagMatch = true
						break
					}
				}
				if !tagMatch {
					matches = false
				}
			}
		}

		// If it passed all filters, add it to the result
		if matches {
			filteredTasks = append(filteredTasks, task)
		}
	}

	return filteredTasks
}

// SortTasks sorts a slice of tasks according to the given sort options
func SortTasks(tasks []*models.Task, options SortOptions) []*models.Task {
	// Make a copy to avoid modifying the original slice
	sortedTasks := make([]*models.Task, len(tasks))
	copy(sortedTasks, tasks)

	// Define sort function based on options
	sorter := func(i, j int) bool {
		taskI := sortedTasks[i]
		taskJ := sortedTasks[j]

		// Determine sort order
		ascending := options.Direction != "desc"

		// Sort by specified field
		switch options.Field {
		case "title":
			if ascending {
				return taskI.Title < taskJ.Title
			}
			return taskI.Title > taskJ.Title
		case "dueDate":
			// Handle nil due dates (put them at the end)
			if taskI.DueDate.IsZero() && !taskJ.DueDate.IsZero() {
				return !ascending
			}
			if !taskI.DueDate.IsZero() && taskJ.DueDate.IsZero() {
				return ascending
			}
			if taskI.DueDate.IsZero() && taskJ.DueDate.IsZero() {
				return false
			}

			if ascending {
				return taskI.DueDate.Before(taskJ.DueDate)
			}
			return taskI.DueDate.After(taskJ.DueDate)
		case "createdAt":
			if ascending {
				return taskI.CreatedAt.Before(taskJ.CreatedAt)
			}
			return taskI.CreatedAt.After(taskJ.CreatedAt)
		case "priority":
			// Convert priority to numeric value for sorting
			priorityValue := map[models.Priority]int{
				models.Low:    1,
				models.Medium: 2,
				models.High:   3,
			}
			valI := priorityValue[taskI.Priority]
			valJ := priorityValue[taskJ.Priority]

			if ascending {
				return valI < valJ
			}
			return valI > valJ
		case "completed":
			if ascending {
				return !taskI.Completed && taskJ.Completed
			}
			return taskI.Completed && !taskJ.Completed
		case "category":
			if ascending {
				return taskI.Category < taskJ.Category
			}
			return taskI.Category > taskJ.Category
		default:
			// Default to sorting by ID
			if ascending {
				return taskI.ID < taskJ.ID
			}
			return taskI.ID > taskJ.ID
		}
	}

	// Sort the slice
	sort.SliceStable(sortedTasks, sorter)

	return sortedTasks
}

// Helper function for string contains search
func Contains(s, substr string) bool {
	return strings.Contains(
		strings.ToLower(s),
		strings.ToLower(substr),
	)
}
