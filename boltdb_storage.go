package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

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
func (s *BoltDBStorage) Add(task *Task) error {
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
		if err := b.Put(itob(task.ID), encoded); err != nil {
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
func (s *BoltDBStorage) Get(id int) (*Task, error) {
	var task *Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)
		v := b.Get(itob(id))
		if v == nil {
			return nil // Not found, but not an error
		}

		// Deserialize the task
		var t Task
		if err := json.Unmarshal(v, &t); err != nil {
			return err
		}
		task = &t

		return nil
	})

	return task, err
}

// GetAll returns all tasks
func (s *BoltDBStorage) GetAll() ([]*Task, error) {
	var tasks []*Task

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task Task
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
func (s *BoltDBStorage) Update(task *Task) error {
	return s.db.Update(func(tx *bolt.Tx) error {
		// Get the old task first for updating indexes
		b := tx.Bucket(taskBucket)
		oldData := b.Get(itob(task.ID))
		if oldData == nil {
			return fmt.Errorf("task not found: %d", task.ID)
		}

		var oldTask Task
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
		if err := b.Put(itob(task.ID), encoded); err != nil {
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
		oldData := b.Get(itob(id))
		if oldData == nil {
			return nil // Task not found, just ignore
		}

		var oldTask Task
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
		return b.Delete(itob(id))
	})
}

// Search finds tasks based on a query string
func (s *BoltDBStorage) Search(query string) ([]*Task, error) {
	var tasks []*Task
	query = strings.ToLower(query)

	err := s.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket(taskBucket)

		return b.ForEach(func(k, v []byte) error {
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			// Check if title or description contains the query
			if strings.Contains(strings.ToLower(task.Title), query) ||
				strings.Contains(strings.ToLower(task.Description), query) {
				tasks = append(tasks, &task)
			}
			return nil
		})
	})

	return tasks, err
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

		// Count all tasks
		var total int
		err := b.ForEach(func(k, v []byte) error {
			total++
			var task Task
			if err := json.Unmarshal(v, &task); err != nil {
				return err
			}

			// Count by completion status
			if task.Completed {
				stats["completed"] = stats["completed"].(int) + 1
			} else {
				stats["pending"] = stats["pending"].(int) + 1
				// Check if overdue
				if !task.DueDate.IsZero() && task.DueDate.Before(time.Now()) {
					stats["overdue"] = stats["overdue"].(int) + 1
				}
			}

			// Count by category
			if task.Category != "" {
				categoryStats := stats["byCategory"].(map[string]int)
				categoryStats[task.Category]++
			}

			// Count by priority
			priorityStats := stats["byPriority"].(map[string]int)
			priorityStats[string(task.Priority)]++

			return nil
		})
		if err != nil {
			return err
		}

		stats["total"] = total
		return nil
	})

	return stats, err
}

// Close closes the database connection
func (s *BoltDBStorage) Close() error {
	return s.db.Close()
}

// Filter filters tasks based on the provided options
func (s *BoltDBStorage) Filter(options FilterOptions) ([]*Task, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	
	return FilterTasks(tasks, options), nil
}

// Sort sorts tasks based on the provided options
func (s *BoltDBStorage) Sort(tasks []*Task, options SortOptions) []*Task {
	return SortTasks(tasks, options)
}

// GetCategories returns all unique categories
func (s *BoltDBStorage) GetCategories() ([]string, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	
	categories := make(map[string]bool)
	for _, task := range tasks {
		if task.Category != "" {
			categories[task.Category] = true
		}
	}
	
	uniqueCategories := make([]string, 0, len(categories))
	for category := range categories {
		uniqueCategories = append(uniqueCategories, category)
	}
	
	return uniqueCategories, nil
}

// GetTags returns all unique tags
func (s *BoltDBStorage) GetTags() ([]string, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	
	tags := make(map[string]bool)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			tags[tag] = true
		}
	}
	
	uniqueTags := make([]string, 0, len(tags))
	for tag := range tags {
		uniqueTags = append(uniqueTags, tag)
	}
	
	return uniqueTags, nil
}

// SaveAttachment saves an attachment for a task
func (s *BoltDBStorage) SaveAttachment(taskID int, filename string, data []byte, contentType string) (*Attachment, error) {
	// Get the task first
	task, err := s.Get(taskID)
	if err != nil {
		return nil, err
	}
	if task == nil {
		return nil, fmt.Errorf("task not found")
	}
	
	// Create a unique path for the attachment
	path := fmt.Sprintf("attachments/%d/%d_%s", taskID, time.Now().Unix(), filename)
	
	// Create the attachment record
	attachment := Attachment{
		ID:          len(task.Attachments) + 1,
		Filename:    filename,
		ContentType: contentType,
		Size:        int64(len(data)),
		UploadedAt:  time.Now(),
		Path:        path,
	}
	
	// Add the attachment to the task
	task.Attachments = append(task.Attachments, attachment)
	
	// Save the task with the new attachment metadata
	err = s.Update(task)
	if err != nil {
		return nil, err
	}
	
	// Save the attachment data to a separate bucket
	err = s.db.Update(func(tx *bolt.Tx) error {
		attachmentsBucket, err := tx.CreateBucketIfNotExists([]byte("attachments"))
		if err != nil {
			return err
		}
		
		// Use taskID_attachmentID as the key
		key := fmt.Sprintf("%d_%d", taskID, attachment.ID)
		
		return attachmentsBucket.Put([]byte(key), data)
	})
	
	if err != nil {
		return nil, err
	}
	
	return &attachment, nil
}

// GetAttachment retrieves an attachment for a task
func (s *BoltDBStorage) GetAttachment(taskID int, attachmentID int) (*Attachment, []byte, error) {
	// Get the task first to get attachment metadata
	task, err := s.Get(taskID)
	if err != nil {
		return nil, nil, err
	}
	if task == nil {
		return nil, nil, fmt.Errorf("task not found")
	}
	
	// Find the attachment metadata
	var attachmentMeta *Attachment
	for _, a := range task.Attachments {
		if a.ID == attachmentID {
			attachmentMeta = &a
			break
		}
	}
	
	if attachmentMeta == nil {
		return nil, nil, fmt.Errorf("attachment not found")
	}
	
	// Get the attachment data from the attachments bucket
	var data []byte
	err = s.db.View(func(tx *bolt.Tx) error {
		attachmentsBucket := tx.Bucket([]byte("attachments"))
		if attachmentsBucket == nil {
			return fmt.Errorf("attachments bucket not found")
		}
		
		// Use taskID_attachmentID as the key
		key := fmt.Sprintf("%d_%d", taskID, attachmentID)
		
		data = attachmentsBucket.Get([]byte(key))
		if data == nil {
			return fmt.Errorf("attachment data not found")
		}
		
		return nil
	})
	
	if err != nil {
		return nil, nil, err
	}
	
	return attachmentMeta, data, nil
}

// DeleteAttachment deletes an attachment from a task
func (s *BoltDBStorage) DeleteAttachment(taskID int, attachmentID int) error {
	// Get the task first
	task, err := s.Get(taskID)
	if err != nil {
		return err
	}
	if task == nil {
		return fmt.Errorf("task not found")
	}
	
	// Find and remove the attachment from the task
	found := false
	for i, a := range task.Attachments {
		if a.ID == attachmentID {
			task.Attachments = append(task.Attachments[:i], task.Attachments[i+1:]...)
			found = true
			break
		}
	}
	
	if !found {
		return fmt.Errorf("attachment not found")
	}
	
	// Update the task without the attachment
	err = s.Update(task)
	if err != nil {
		return err
	}
	
	// Remove the attachment data from the attachments bucket
	return s.db.Update(func(tx *bolt.Tx) error {
		attachmentsBucket := tx.Bucket([]byte("attachments"))
		if attachmentsBucket == nil {
			return nil // Bucket doesn't exist, nothing to delete
		}
		
		// Use taskID_attachmentID as the key
		key := fmt.Sprintf("%d_%d", taskID, attachmentID)
		
		return attachmentsBucket.Delete([]byte(key))
	})
}

// GetTasksByUser returns tasks created by or assigned to a specific user
func (s *BoltDBStorage) GetTasksByUser(username string) ([]*Task, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	
	userTasks := make([]*Task, 0)
	for _, task := range tasks {
		if task.CreatedBy == username || task.AssignedTo == username {
			userTasks = append(userTasks, task)
		}
	}
	
	return userTasks, nil
}

// GetSharedTasks returns tasks shared with a specific user
func (s *BoltDBStorage) GetSharedTasks(username string) ([]*Task, error) {
	tasks, err := s.GetAll()
	if err != nil {
		return nil, err
	}
	
	sharedTasks := make([]*Task, 0)
	for _, task := range tasks {
		for _, sharedWith := range task.SharedWith {
			if sharedWith == username {
				sharedTasks = append(sharedTasks, task)
				break
			}
		}
	}
	
	return sharedTasks, nil
}

// Helper functions


// indexTags adds tag indexes for a task
func indexTags(tx *bolt.Tx, task *Task) error {
	b := tx.Bucket(tagsBucket)
	for _, tag := range task.Tags {
		tagKey := []byte(tag)
		var taskIDs []int

		// Get existing task IDs for this tag
		existing := b.Get(tagKey)
		if existing != nil {
			if err := json.Unmarshal(existing, &taskIDs); err != nil {
				return err
			}
		}

		// Add the current task ID if not already in the list
		found := false
		for _, id := range taskIDs {
			if id == task.ID {
				found = true
				break
			}
		}

		if !found {
			taskIDs = append(taskIDs, task.ID)
			encoded, err := json.Marshal(taskIDs)
			if err != nil {
				return err
			}

			if err := b.Put(tagKey, encoded); err != nil {
				return err
			}
		}
	}
	return nil
}

// removeTagsIndex removes tag indexes for a task
func removeTagsIndex(tx *bolt.Tx, task *Task) error {
	b := tx.Bucket(tagsBucket)
	for _, tag := range task.Tags {
		tagKey := []byte(tag)
		var taskIDs []int

		// Get existing task IDs for this tag
		existing := b.Get(tagKey)
		if existing != nil {
			if err := json.Unmarshal(existing, &taskIDs); err != nil {
				return err
			}

			// Remove the task ID from the list
			var newIDs []int
			for _, id := range taskIDs {
				if id != task.ID {
					newIDs = append(newIDs, id)
				}
			}

			// Update or delete the tag entry
			if len(newIDs) > 0 {
				encoded, err := json.Marshal(newIDs)
				if err != nil {
					return err
				}

				if err := b.Put(tagKey, encoded); err != nil {
					return err
				}
			} else {
				if err := b.Delete(tagKey); err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// indexCategory adds category indexes for a task
func indexCategory(tx *bolt.Tx, task *Task) error {
	if task.Category == "" {
		return nil
	}

	b := tx.Bucket(categoriesBucket)
	categoryKey := []byte(task.Category)
	var taskIDs []int

	// Get existing task IDs for this category
	existing := b.Get(categoryKey)
	if existing != nil {
		if err := json.Unmarshal(existing, &taskIDs); err != nil {
			return err
		}
	}

	// Add the current task ID if not already in the list
	found := false
	for _, id := range taskIDs {
		if id == task.ID {
			found = true
			break
		}
	}

	if !found {
		taskIDs = append(taskIDs, task.ID)
		encoded, err := json.Marshal(taskIDs)
		if err != nil {
			return err
		}

		if err := b.Put(categoryKey, encoded); err != nil {
			return err
		}
	}
	return nil
}

// removeCategoryIndex removes category indexes for a task
func removeCategoryIndex(tx *bolt.Tx, task *Task) error {
	if task.Category == "" {
		return nil
	}

	b := tx.Bucket(categoriesBucket)
	categoryKey := []byte(task.Category)
	var taskIDs []int

	// Get existing task IDs for this category
	existing := b.Get(categoryKey)
	if existing != nil {
		if err := json.Unmarshal(existing, &taskIDs); err != nil {
			return err
		}

		// Remove the task ID from the list
		var newIDs []int
		for _, id := range taskIDs {
			if id != task.ID {
				newIDs = append(newIDs, id)
			}
		}

		// Update or delete the category entry
		if len(newIDs) > 0 {
			encoded, err := json.Marshal(newIDs)
			if err != nil {
				return err
			}

			if err := b.Put(categoryKey, encoded); err != nil {
				return err
			}
		} else {
			if err := b.Delete(categoryKey); err != nil {
				return err
			}
		}
	}
	return nil
}
