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

// Close closes the database
func (s *BoltDBStorage) Close() error {
	return s.db.Close()
}

// Helper functions

// itob converts an int to a byte slice
func itob(v int) []byte {
	key := fmt.Sprintf("%d", v)
	return []byte(key)
}

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
