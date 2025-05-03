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
	DaysOfWeek     []int     // for weekly pattern (0=Sunday, 6=Saturday)
	DayOfMonth     int       // for monthly pattern
	MonthOfYear    int       // for yearly pattern
}

// Reminder represents a notification for a task
type Reminder struct {
	ID          int       `json:"id"`
	Time        time.Time `json:"time"`        // When to send the reminder
	Description string    `json:"description"` // Optional custom message
	Sent        bool      `json:"sent"`        // Whether the reminder has been sent
	Method      string    `json:"method"`      // Email, push, etc.
}

// Note represents a comment or note on a task
type Note struct {
	ID        int
	Content   string
	CreatedAt time.Time
	UpdatedAt time.Time
	Author    string // who created the note
}

// Attachment represents a file attached to a task
type Attachment struct {
	ID          int       `json:"id"`
	Filename    string    `json:"filename"`
	ContentType string    `json:"contentType"`
	Size        int64     `json:"size"`
	UploadedAt  time.Time `json:"uploadedAt"`
	Path        string    `json:"path"`
}

// Subtask represents a smaller task within a parent task
type Subtask struct {
	ID          int       `json:"id"`
	Title       string    `json:"title"`
	Completed   bool      `json:"completed"`
	CreatedAt   time.Time `json:"createdAt"`
	CompletedAt time.Time `json:"completedAt,omitempty"`
}

// Task represents a task in our task manager
type Task struct {
	ID          int          `json:"id"`
	Title       string       `json:"title"`
	Description string       `json:"description"`
	Completed   bool         `json:"completed"`
	CreatedAt   time.Time    `json:"createdAt"`
	DueDate     time.Time    `json:"dueDate"`
	CompletedAt time.Time    `json:"completedAt,omitempty"`
	Category    string       `json:"category"`
	Priority    Priority     `json:"priority"`
	Tags        []string     `json:"tags"`
	Notes       []Note       `json:"notes"`
	Recurrence  *Recurrence  `json:"recurrence,omitempty"` // Recurrence pattern (optional)
	Subtasks    []Subtask    `json:"subtasks,omitempty"`
	Attachments []Attachment `json:"attachments,omitempty"`
	AssignedTo  string       `json:"assignedTo,omitempty"` // Username of assignee
	CreatedBy   string       `json:"createdBy,omitempty"`  // Username of creator
	SharedWith  []string     `json:"sharedWith,omitempty"` // Usernames of people with access
	Reminders   []Reminder   `json:"reminders,omitempty"`  // Task reminders
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
		Subtasks:    make([]Subtask, 0),
		Attachments: make([]Attachment, 0),
		SharedWith:  make([]string, 0),
		Reminders:   make([]Reminder, 0),
	}
}

// AddNote adds a new note to the task
func (t *Task) AddNote(content string, author string) {
	note := Note{
		ID:        len(t.Notes) + 1,
		Content:   content,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Author:    author,
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

// AddSubtask adds a new subtask to the task
func (t *Task) AddSubtask(title string) {
	subtask := Subtask{
		ID:        len(t.Subtasks) + 1,
		Title:     title,
		Completed: false,
		CreatedAt: time.Now(),
	}
	t.Subtasks = append(t.Subtasks, subtask)
}

// CompleteSubtask marks a subtask as completed
func (t *Task) CompleteSubtask(subtaskID int) bool {
	for i, subtask := range t.Subtasks {
		if subtask.ID == subtaskID {
			t.Subtasks[i].Completed = true
			t.Subtasks[i].CompletedAt = time.Now()
			return true
		}
	}
	return false
}

// ReopenSubtask marks a subtask as not completed
func (t *Task) ReopenSubtask(subtaskID int) bool {
	for i, subtask := range t.Subtasks {
		if subtask.ID == subtaskID {
			t.Subtasks[i].Completed = false
			t.Subtasks[i].CompletedAt = time.Time{}
			return true
		}
	}
	return false
}

// DeleteSubtask removes a subtask from the task
func (t *Task) DeleteSubtask(subtaskID int) bool {
	for i, subtask := range t.Subtasks {
		if subtask.ID == subtaskID {
			t.Subtasks = append(t.Subtasks[:i], t.Subtasks[i+1:]...)
			return true
		}
	}
	return false
}

// AddAttachment adds a new attachment to the task
func (t *Task) AddAttachment(filename, contentType, path string, size int64) {
	attachment := Attachment{
		ID:          len(t.Attachments) + 1,
		Filename:    filename,
		ContentType: contentType,
		Size:        size,
		UploadedAt:  time.Now(),
		Path:        path,
	}
	t.Attachments = append(t.Attachments, attachment)
}

// RemoveAttachment removes an attachment from the task
func (t *Task) RemoveAttachment(attachmentID int) bool {
	for i, attachment := range t.Attachments {
		if attachment.ID == attachmentID {
			t.Attachments = append(t.Attachments[:i], t.Attachments[i+1:]...)
			return true
		}
	}
	return false
}

// ShareWith shares the task with another user
func (t *Task) ShareWith(username string) {
	// Check if already shared with this user
	for _, shared := range t.SharedWith {
		if shared == username {
			return
		}
	}
	t.SharedWith = append(t.SharedWith, username)
}

// UnshareWith removes sharing with a user
func (t *Task) UnshareWith(username string) bool {
	for i, shared := range t.SharedWith {
		if shared == username {
			t.SharedWith = append(t.SharedWith[:i], t.SharedWith[i+1:]...)
			return true
		}
	}
	return false
}

// MarkCompleted marks the task as completed
func (t *Task) MarkCompleted() {
	t.Completed = true
	t.CompletedAt = time.Now()
}

// MarkIncomplete marks the task as not completed
func (t *Task) MarkIncomplete() {
	t.Completed = false
	t.CompletedAt = time.Time{}
}

// AddReminder adds a new reminder to the task
func (t *Task) AddReminder(reminderTime time.Time, description, method string) {
	reminder := Reminder{
		ID:          len(t.Reminders) + 1,
		Time:        reminderTime,
		Description: description,
		Sent:        false,
		Method:      method,
	}
	t.Reminders = append(t.Reminders, reminder)
}

// RemoveReminder removes a reminder from the task
func (t *Task) RemoveReminder(reminderID int) bool {
	for i, reminder := range t.Reminders {
		if reminder.ID == reminderID {
			t.Reminders = append(t.Reminders[:i], t.Reminders[i+1:]...)
			return true
		}
	}
	return false
}

// MarkReminderSent marks a reminder as sent
func (t *Task) MarkReminderSent(reminderID int) bool {
	for i, reminder := range t.Reminders {
		if reminder.ID == reminderID {
			t.Reminders[i].Sent = true
			return true
		}
	}
	return false
}

// Simple ID generator (in a real app, this would be more sophisticated)
func generateID() int {
	return int(time.Now().Unix())
}
