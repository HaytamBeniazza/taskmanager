package main

import (
	"time"
)

// Priority represents task importance level
type Priority string

const (
	Low    Priority = "low"
	Medium Priority = "medium"
	High   Priority = "high"
)

// Recurrence represents task recurrence pattern
type Recurrence struct {
	Pattern        string    // daily, weekly, monthly, yearly
	Interval       int       // interval between occurrences
	EndDate        time.Time // when to stop recurring
	LastOccurrence time.Time // when the task last occurred
}

// Note represents a comment or note on a task
type Note struct {
	ID        int
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Task represents a task in our task manager
type Task struct {
	ID          int         `json:"id"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Completed   bool        `json:"completed"`
	CreatedAt   time.Time   `json:"createdAt"`
	DueDate     time.Time   `json:"dueDate"`
	Category    string      `json:"category"`
	Priority    Priority    `json:"priority"`
	Tags        []string    `json:"tags"`
	Notes       []Note      `json:"notes"`
	Recurrence  *Recurrence // Recurrence pattern (optional)
}

// NewTask creates a new task with the given title and description
func NewTask(title, description string) *Task {
	return &Task{
		Title:       title,
		Description: description,
		CreatedAt:   time.Now(),
		Priority:    Medium,
		Tags:        make([]string, 0),
		Notes:       make([]Note, 0),
	}
}

// AddNote adds a new note to the task
func (t *Task) AddNote(content string) {
	note := Note{
		ID:        len(t.Notes) + 1,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	t.Notes = append(t.Notes, note)
}

// UpdateNote updates an existing note
func (t *Task) UpdateNote(noteID int, content string) bool {
	for i, note := range t.Notes {
		if note.ID == noteID {
			t.Notes[i].Content = content
			t.Notes[i].UpdatedAt = time.Now()
			return true
		}
	}
	return false
}

// DeleteNote removes a note from the task
func (t *Task) DeleteNote(noteID int) bool {
	for i, note := range t.Notes {
		if note.ID == noteID {
			t.Notes = append(t.Notes[:i], t.Notes[i+1:]...)
			return true
		}
	}
	return false
}

// AddTag adds a new tag to the task
func (t *Task) AddTag(tag string) {
	t.Tags = append(t.Tags, tag)
}

// RemoveTag removes a tag from the task
func (t *Task) RemoveTag(tag string) bool {
	for i, existingTag := range t.Tags {
		if existingTag == tag {
			t.Tags = append(t.Tags[:i], t.Tags[i+1:]...)
			return true
		}
	}
	return false
}

// Simple ID generator (in a real app, this would be more sophisticated)
func generateID() int {
	return int(time.Now().Unix())
}
