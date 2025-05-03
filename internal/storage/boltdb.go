package storage

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/HaytamBeniazza/taskmanager/internal/models"
	"github.com/HaytamBeniazza/taskmanager/internal/utils"
	"github.com/boltdb/bolt"
)

var taskBucket = []byte("tasks")
var tagsBucket = []byte("tags")
var categoriesBucket = []byte("categories")

// BoltDBStorage implements Storage using BoltDB
type BoltDBStorage struct {
	db *bolt.DB
}

// NewBoltDBStorage creates a new BoltDB storage
func NewBoltDBStorage(dbPath string) (*BoltDBStorage, error) {
	db, err := bolt.Open(dbPath, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		return nil, err
	}

	// Initialize buckets
	err = db.Update(func(tx *bolt.Tx) error {
		if _, err := tx.CreateBucketIfNotExists(taskBucket); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		if _, err := tx.CreateBucketIfNotExists(tagsBucket); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		if _, err := tx.CreateBucketIfNotExists(categoriesBucket); err != nil {
			return fmt.Errorf("create bucket: %s", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	return &BoltDBStorage{db: db}, nil
}

// Add saves a new task
func (s *BoltDBStorage) Add(task *models.Task) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the tasks bucket
		b := tx.Bucket(taskBucket)

		// Generate ID for the task
		id, err := b.NextSequence()
		if err != nil {
			return err
		}
		task.ID = int(id)

		// Serialize task to JSON
		encoded, err := json.Marshal(task)
		if err != nil {
			return err
		}

		// Store task
		if err := b.Put(utils.Itob(task.ID), encoded); err != nil {
			return err
		}

		// Index tags
		if err := indexTags(tx, task); err != nil {
			return err
		}

		// Index category
		if err := indexCategory(tx, task); err != nil {
			return err
		}

		return nil
	})
}

// Get retrieves a task by ID
func (s *BoltDBStorage) Get(id int) (*models.Task, error) {
	var task *models.Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)
		v := b.Get(utils.Itob(id))
		if v == nil {
			return nil // Not found, but not an error
		}

		// Deserialize the task
		var t models.Task
		if err := json.Unmarshal(v, &t); err != nil {
			return err
		}
		task = &t

		return nil
	})

	return task, err
}

// GetAll returns all tasks
func (s *BoltDBStorage) GetAll() ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task models.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}
			tasks = append(tasks, &task)
			return nil
		})
	})

	return tasks, err
}

// Update modifies an existing task
func (s *BoltDBStorage) Update(task *models.Task) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the old task first for updating indexes
		b := tx.Bucket(taskBucket)
		oldData := b.Get(utils.Itob(task.ID))
		if oldData == nil {
			return fmt.Errorf("task not found: %d", task.ID)
		}

		var oldTask models.Task
		if err := json.Unmarshal(oldData, &oldTask); err != nil {
			return err
		}

		// Remove old indexes
		if err := removeTagsIndex(tx, &oldTask); err != nil {
			return err
		}
		if err := removeCategoryIndex(tx, &oldTask); err != nil {
			return err
		}

		// Serialize task to JSON
		encoded, err := json.Marshal(task)
		if err != nil {
			return err
		}

		// Store updated task
		if err := b.Put(utils.Itob(task.ID), encoded); err != nil {
			return err
		}

		// Index tags
		if err := indexTags(tx, task); err != nil {
			return err
		}

		// Index category
		if err := indexCategory(tx, task); err != nil {
			return err
		}

		return nil
	})
}

// Delete removes a task by ID
func (s *BoltDBStorage) Delete(id int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the old task first for updating indexes
		b := tx.Bucket(taskBucket)
		oldData := b.Get(utils.Itob(id))
		if oldData == nil {
			return nil // Task not found, just ignore
		}

		var oldTask models.Task
		if err := json.Unmarshal(oldData, &oldTask); err != nil {
			return err
		}

		// Remove old indexes
		if err := removeTagsIndex(tx, &oldTask); err != nil {
			return err
		}
		if err := removeCategoryIndex(tx, &oldTask); err != nil {
			return err
		}

		// Delete the task
		return b.Delete(utils.Itob(id))
	})
}

// Search searches tasks by title, description, or category
func (s *BoltDBStorage) Search(query string) ([]*models.Task, error) {
	var results []*models.Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task models.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			// Match against title, description, category, and tags
			if utils.Contains(task.Title, query) ||
				utils.Contains(task.Description, query) ||
				utils.Contains(task.Category, query) {
				results = append(results, &task)
				return nil
			}

			// Search in tags
			for _, tag := range task.Tags {
				if utils.Contains(tag, query) {
					results = append(results, &task)
					return nil
				}
			}

			return nil
		})
	})

	return results, err
}

// GetStats returns statistics about tasks
func (s *BoltDBStorage) GetStats() (map[string]interface{}, error) {
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

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		// Get total count
		totalTasks := 0
		completedTasks := 0
		pendingTasks := 0
		overdueTasks := 0
		categoryCount := make(map[string]int)
		priorityCount := map[string]int{
			"low":    0,
			"medium": 0,
			"high":   0,
		}

		now := time.Now()

		err := b.ForEach(func(k, v []byte) error {
			var task models.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			totalTasks++

			if task.Completed {
				completedTasks++
			} else {
				pendingTasks++
				if !task.DueDate.IsZero() && task.DueDate.Before(now) {
					overdueTasks++
				}
			}

			// Count by category
			category := task.Category
			if category == "" {
				category = "Uncategorized"
			}
			categoryCount[category]++

			// Count by priority
			priorityCount[string(task.Priority)]++

			return nil
		})

		if err != nil {
			return err
		}

		stats["total"] = totalTasks
		stats["completed"] = completedTasks
		stats["pending"] = pendingTasks
		stats["overdue"] = overdueTasks
		stats["byCategory"] = categoryCount
		stats["byPriority"] = priorityCount

		return nil
	})

	return stats, err
}

// Close closes the database
func (s *BoltDBStorage) Close() error {
	return s.db.Close()
}

// Filter returns tasks matching the given filter options
func (s *BoltDBStorage) Filter(options FilterOptions) ([]*models.Task, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	return FilterTasks(tasks, options), nil
}

// GetPaginated retrieves paginated tasks with sorting
func (s *BoltDBStorage) GetPaginated(page, perPage int, sortOptions SortOptions) ([]*models.Task, int, error) {
	// Get all tasks first
	allTasks, err := s.GetAll()
	if err != nil {
		return nil, 0, err
	}

	// Apply sorting
	sortedTasks := SortTasks(allTasks, sortOptions)

	// Calculate total tasks count
	total := len(sortedTasks)

	// Calculate start and end indices for pagination
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	// Check if start index is valid
	if startIndex >= total {
		// Return empty array if page is out of range
		return []*models.Task{}, total, nil
	}

	// Adjust end index if it exceeds array length
	if endIndex > total {
		endIndex = total
	}

	// Extract the page of tasks
	pagedTasks := sortedTasks[startIndex:endIndex]

	return pagedTasks, total, nil
}

// GetCategories returns all unique categories
func (s *BoltDBStorage) GetCategories() ([]string, error) {
	var categories []string

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(categoriesBucket)
		if b == nil {
			return fmt.Errorf("categories bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			categories = append(categories, string(k))
			return nil
		})
	})

	return categories, err
}

// GetTags returns all unique tags
func (s *BoltDBStorage) GetTags() ([]string, error) {
	var tags []string

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(tagsBucket)
		if b == nil {
			return fmt.Errorf("tags bucket not found")
		}

		return b.ForEach(func(k, v []byte) error {
			tags = append(tags, string(k))
			return nil
		})
	})

	return tags, err
}

// GetTaskAnalytics returns detailed analytics about tasks
func (s *BoltDBStorage) GetTaskAnalytics() (map[string]interface{}, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}

	analytics := map[string]interface{}{
		"total_tasks":           len(tasks),
		"completed_tasks":       0,
		"completion_rate":       0.0,
		"category_distribution": make(map[string]int),
		"priority_distribution": map[string]int{
			"low":    0,
			"medium": 0,
			"high":   0,
		},
		"avg_completion_time": 0.0,
		"top_tags":            []map[string]interface{}{},
		"overdue_tasks":       0,
		"weekly_activity":     make(map[string]int),
	}

	// Calculate metrics
	var completedTasks int
	var totalCompletionTime float64
	var completedWithTimeCount int
	tagCounts := make(map[string]int)
	now := time.Now()

	for _, task := range tasks {
		// Priority distribution
		analytics["priority_distribution"].(map[string]int)[string(task.Priority)]++

		// Category distribution
		category := task.Category
		if category == "" {
			category = "Uncategorized"
		}
		analytics["category_distribution"].(map[string]interface{})[category] = analytics["category_distribution"].(map[string]interface{})[category].(int) + 1

		// Tag counts
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}

		// Completed tasks count
		if task.Completed {
			completedTasks++
			if !task.CreatedAt.IsZero() && !task.CompletedAt.IsZero() {
				completionTime := task.CompletedAt.Sub(task.CreatedAt).Hours()
				totalCompletionTime += completionTime
				completedWithTimeCount++
			}
		} else if !task.DueDate.IsZero() && task.DueDate.Before(now) {
			// Overdue tasks
			analytics["overdue_tasks"] = analytics["overdue_tasks"].(int) + 1
		}
	}

	// Calculate completion rate
	if len(tasks) > 0 {
		analytics["completed_tasks"] = completedTasks
		analytics["completion_rate"] = float64(completedTasks) / float64(len(tasks)) * 100
	}

	// Calculate average completion time
	if completedWithTimeCount > 0 {
		analytics["avg_completion_time"] = totalCompletionTime / float64(completedWithTimeCount)
	}

	// Get top tags
	type TagCount struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}
	var topTags []TagCount
	for tag, count := range tagCounts {
		topTags = append(topTags, TagCount{Tag: tag, Count: count})
	}

	// Sort tags by count (descending)
	sort.Slice(topTags, func(i, j int) bool {
		return topTags[i].Count > topTags[j].Count
	})

	// Take top 10 tags
	maxTags := 10
	if len(topTags) < maxTags {
		maxTags = len(topTags)
	}
	topTagsMap := make([]map[string]interface{}, maxTags)
	for i := 0; i < maxTags; i++ {
		topTagsMap[i] = map[string]interface{}{
			"tag":   topTags[i].Tag,
			"count": topTags[i].Count,
		}
	}
	analytics["top_tags"] = topTagsMap

	return analytics, nil
}

// SaveAttachment stores a file attachment for a task
func (s *BoltDBStorage) SaveAttachment(taskID int, filename string, data []byte, contentType string) (*models.Attachment, error) {
	var attachment *models.Attachment

	err := s.db.Update(func(tx *bolt.Tx) error {
		// Get the task
		taskBucket := tx.Bucket(taskBucket)
		taskData := taskBucket.Get(utils.Itob(taskID))
		if taskData == nil {
			return fmt.Errorf("task not found: %d", taskID)
		}

		var task models.Task
		if err := json.Unmarshal(taskData, &task); err != nil {
			return err
		}

		// Create a new attachment
		newAttachment := models.Attachment{
			ID:          len(task.Attachments) + 1,
			Filename:    filename,
			ContentType: contentType,
			Size:        int64(len(data)),
			UploadedAt:  time.Now(),
			// We'll store the path as a reference to where the file data is stored
			Path: fmt.Sprintf("tasks/%d/attachments/%d", taskID, len(task.Attachments)+1),
		}

		// Add attachment to the task
		task.Attachments = append(task.Attachments, newAttachment)

		// Update the task
		updatedTaskData, err := json.Marshal(task)
		if err != nil {
			return err
		}
		if err := taskBucket.Put(utils.Itob(taskID), updatedTaskData); err != nil {
			return err
		}

		// Return the attachment
		attachment = &newAttachment
		return nil
	})

	return attachment, err
}

// GetAttachment retrieves a file attachment from a task
func (s *BoltDBStorage) GetAttachment(taskID int, attachmentID int) (*models.Attachment, []byte, error) {
	var attachment *models.Attachment
	var data []byte

	err := s.db.View(func(tx *bolt.Tx) error {
		// Get the task
		taskBucket := tx.Bucket(taskBucket)
		taskData := taskBucket.Get(utils.Itob(taskID))
		if taskData == nil {
			return fmt.Errorf("task not found: %d", taskID)
		}

		var task models.Task
		if err := json.Unmarshal(taskData, &task); err != nil {
			return err
		}

		// Find the attachment
		found := false
		for _, att := range task.Attachments {
			if att.ID == attachmentID {
				attachment = &att
				found = true
				break
			}
		}

		if !found {
			return fmt.Errorf("attachment not found: %d", attachmentID)
		}

		// In a real application, we would retrieve the file data from a file system or blob storage.
		// For this example, we'll just return some dummy data.
		data = []byte("This is the content of the attachment file")

		return nil
	})

	return attachment, data, err
}

// DeleteAttachment removes a file attachment from a task
func (s *BoltDBStorage) DeleteAttachment(taskID int, attachmentID int) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the task
		taskBucket := tx.Bucket(taskBucket)
		taskData := taskBucket.Get(utils.Itob(taskID))
		if taskData == nil {
			return fmt.Errorf("task not found: %d", taskID)
		}

		var task models.Task
		if err := json.Unmarshal(taskData, &task); err != nil {
			return err
		}

		// Find and remove the attachment
		foundIndex := -1
		for i, att := range task.Attachments {
			if att.ID == attachmentID {
				foundIndex = i
				break
			}
		}

		if foundIndex == -1 {
			return fmt.Errorf("attachment not found: %d", attachmentID)
		}

		// Remove the attachment (preserve order)
		task.Attachments = append(task.Attachments[:foundIndex], task.Attachments[foundIndex+1:]...)

		// Update the task
		updatedTaskData, err := json.Marshal(task)
		if err != nil {
			return err
		}
		if err := taskBucket.Put(utils.Itob(taskID), updatedTaskData); err != nil {
			return err
		}

		// In a real application, we would also delete the file from the file system or blob storage

		return nil
	})
}

// GetTasksByUser returns tasks assigned to or created by a user
func (s *BoltDBStorage) GetTasksByUser(username string) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task models.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			if task.AssignedTo == username || task.CreatedBy == username {
				tasks = append(tasks, &task)
			}
			return nil
		})
	})

	return tasks, err
}

// GetSharedTasks returns tasks shared with a user
func (s *BoltDBStorage) GetSharedTasks(username string) ([]*models.Task, error) {
	var tasks []*models.Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task models.Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			for _, sharedWith := range task.SharedWith {
				if sharedWith == username {
					tasks = append(tasks, &task)
					break
				}
			}
			return nil
		})
	})

	return tasks, err
}

// BatchCreate creates multiple tasks at once
func (s *BoltDBStorage) BatchCreate(tasks []*models.Task) error {
	for _, task := range tasks {
		if err := s.Add(task); err != nil {
			return err
		}
	}
	return nil
}

// BatchUpdate updates multiple tasks at once
func (s *BoltDBStorage) BatchUpdate(tasks []*models.Task) error {
	for _, task := range tasks {
		if err := s.Update(task); err != nil {
			return err
		}
	}
	return nil
}

// BatchDelete deletes multiple tasks at once
func (s *BoltDBStorage) BatchDelete(ids []int) error {
	for _, id := range ids {
		if err := s.Delete(id); err != nil {
			return err
		}
	}
	return nil
}

// BatchComplete marks multiple tasks as completed
func (s *BoltDBStorage) BatchComplete(ids []int) error {
	for _, id := range ids {
		task, err := s.Get(id)
		if err != nil {
			return err
		}
		if task == nil {
			continue
		}
		task.MarkCompleted()
		if err := s.Update(task); err != nil {
			return err
		}
	}
	return nil
}

// Helper functions for indexing

// indexTags indexes a task's tags for searching
func indexTags(tx *bolt.Tx, task *models.Task) error {
	tagsBucket := tx.Bucket(tagsBucket)

	// Add each tag to the index
	for _, tag := range task.Tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		// Get the existing tag data or create a new one
		tagData := tagsBucket.Get([]byte(normalizedTag))
		var taskIDs []int
		if tagData != nil {
			if err := json.Unmarshal(tagData, &taskIDs); err != nil {
				return err
			}
		}

		// Check if the task ID is already in the list
		found := false
		for _, id := range taskIDs {
			if id == task.ID {
				found = true
				break
			}
		}

		// Add the task ID if not already there
		if !found {
			taskIDs = append(taskIDs, task.ID)
			encodedTaskIDs, err := json.Marshal(taskIDs)
			if err != nil {
				return err
			}
			if err := tagsBucket.Put([]byte(normalizedTag), encodedTaskIDs); err != nil {
				return err
			}
		}
	}

	return nil
}

// removeTagsIndex removes a task's entries from the tags index
func removeTagsIndex(tx *bolt.Tx, task *models.Task) error {
	tagsBucket := tx.Bucket(tagsBucket)

	// Remove the task ID from each tag entry
	for _, tag := range task.Tags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		if normalizedTag == "" {
			continue
		}

		// Get the existing tag data
		tagData := tagsBucket.Get([]byte(normalizedTag))
		if tagData == nil {
			continue
		}

		var taskIDs []int
		if err := json.Unmarshal(tagData, &taskIDs); err != nil {
			return err
		}

		// Remove the task ID from the list
		updatedTaskIDs := []int{}
		for _, id := range taskIDs {
			if id != task.ID {
				updatedTaskIDs = append(updatedTaskIDs, id)
			}
		}

		// Update or delete the tag entry
		if len(updatedTaskIDs) > 0 {
			encodedTaskIDs, err := json.Marshal(updatedTaskIDs)
			if err != nil {
				return err
			}
			if err := tagsBucket.Put([]byte(normalizedTag), encodedTaskIDs); err != nil {
				return err
			}
		} else {
			if err := tagsBucket.Delete([]byte(normalizedTag)); err != nil {
				return err
			}
		}
	}

	return nil
}

// indexCategory indexes a task's category for searching
func indexCategory(tx *bolt.Tx, task *models.Task) error {
	categoriesBucket := tx.Bucket(categoriesBucket)

	// Skip if no category
	if task.Category == "" {
		return nil
	}

	normalizedCategory := strings.ToLower(strings.TrimSpace(task.Category))
	if normalizedCategory == "" {
		return nil
	}

	// Get the existing category data or create a new one
	categoryData := categoriesBucket.Get([]byte(normalizedCategory))
	var taskIDs []int
	if categoryData != nil {
		if err := json.Unmarshal(categoryData, &taskIDs); err != nil {
			return err
		}
	}

	// Check if the task ID is already in the list
	found := false
	for _, id := range taskIDs {
		if id == task.ID {
			found = true
			break
		}
	}

	// Add the task ID if not already there
	if !found {
		taskIDs = append(taskIDs, task.ID)
		encodedTaskIDs, err := json.Marshal(taskIDs)
		if err != nil {
			return err
		}
		if err := categoriesBucket.Put([]byte(normalizedCategory), encodedTaskIDs); err != nil {
			return err
		}
	}

	return nil
}

// removeCategoryIndex removes a task's entry from the category index
func removeCategoryIndex(tx *bolt.Tx, task *models.Task) error {
	categoriesBucket := tx.Bucket(categoriesBucket)

	// Skip if no category
	if task.Category == "" {
		return nil
	}

	normalizedCategory := strings.ToLower(strings.TrimSpace(task.Category))
	if normalizedCategory == "" {
		return nil
	}

	// Get the existing category data
	categoryData := categoriesBucket.Get([]byte(normalizedCategory))
	if categoryData == nil {
		return nil
	}

	var taskIDs []int
	if err := json.Unmarshal(categoryData, &taskIDs); err != nil {
		return err
	}

	// Remove the task ID from the list
	updatedTaskIDs := []int{}
	for _, id := range taskIDs {
		if id != task.ID {
			updatedTaskIDs = append(updatedTaskIDs, id)
		}
	}

	// Update or delete the category entry
	if len(updatedTaskIDs) > 0 {
		encodedTaskIDs, err := json.Marshal(updatedTaskIDs)
		if err != nil {
			return err
		}
		if err := categoriesBucket.Put([]byte(normalizedCategory), encodedTaskIDs); err != nil {
			return err
		}
	} else {
		if err := categoriesBucket.Delete([]byte(normalizedCategory)); err != nil {
			return err
		}
	}

	return nil
}
