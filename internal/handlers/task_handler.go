package handlers

import (
	"net/http"
	"strconv"
	"time"

	"github.com/HaytamBeniazza/taskmanager/internal/models"
	"github.com/HaytamBeniazza/taskmanager/internal/storage"
	"github.com/gin-gonic/gin"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	storage storage.TaskStorage
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(storage storage.TaskStorage) *TaskHandler {
	return &TaskHandler{
		storage: storage,
	}
}

// GetTasks returns all tasks with optional filtering and sorting
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// Parse pagination parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Parse sort options
	sortOptions := storage.SortOptions{
		Field:     c.DefaultQuery("sortBy", "createdAt"),
		Direction: c.DefaultQuery("sortDir", "desc"),
	}

	// Get tasks with pagination and sorting
	tasks, total, err := h.storage.GetPaginated(page, perPage, sortOptions)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate total pages
	totalPages := (total + perPage - 1) / perPage

	// Construct the response with pagination info
	c.JSON(http.StatusOK, gin.H{
		"tasks": tasks,
		"pagination": gin.H{
			"total":       total,
			"currentPage": page,
			"perPage":     perPage,
			"totalPages":  totalPages,
			"hasMore":     page < totalPages,
		},
	})
}

// SearchTasks searches tasks by title, description, or category
func (h *TaskHandler) SearchTasks(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Search query is required"})
		return
	}

	tasks, err := h.storage.Search(query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// GetTask returns a specific task by ID
func (h *TaskHandler) GetTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.Add(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"task": task})
}

// UpdateTask updates an existing task
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task models.Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = id
	if err := h.storage.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// DeleteTask deletes a task
func (h *TaskHandler) DeleteTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	if err := h.storage.Delete(id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// HealthCheck provides a simple health check endpoint
func (h *TaskHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "up",
		"time":   time.Now(),
	})
}

// GetCategories returns all available task categories
func (h *TaskHandler) GetCategories(c *gin.Context) {
	categories, err := h.storage.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"categories": categories})
}

// GetTags returns all available task tags
func (h *TaskHandler) GetTags(c *gin.Context) {
	tags, err := h.storage.GetTags()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tags": tags})
}

// GetTaskStats returns statistics about tasks
func (h *TaskHandler) GetTaskStats(c *gin.Context) {
	stats, err := h.storage.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, stats)
}

// GetTaskAnalytics returns more detailed analytics about tasks
func (h *TaskHandler) GetTaskAnalytics(c *gin.Context) {
	analytics, err := h.storage.GetTaskAnalytics()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, analytics)
}

// AddTaskNote adds a note to a task
func (h *TaskHandler) AddTaskNote(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var noteData struct {
		Content string `json:"content" binding:"required"`
		Author  string `json:"author"`
	}

	if err := c.ShouldBindJSON(&noteData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task.AddNote(noteData.Content, noteData.Author)

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// UpdateTaskNote updates a task note
func (h *TaskHandler) UpdateTaskNote(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	noteID, err := strconv.Atoi(c.Param("noteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	var noteData struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&noteData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.UpdateNote(noteID, noteData.Content) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// DeleteTaskNote deletes a note from a task
func (h *TaskHandler) DeleteTaskNote(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	noteID, err := strconv.Atoi(c.Param("noteId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.DeleteNote(noteID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// AddTaskTag adds a tag to a task
func (h *TaskHandler) AddTaskTag(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var tagData struct {
		Tag string `json:"tag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tagData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task.AddTag(tagData.Tag)

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// RemoveTaskTag removes a tag from a task
func (h *TaskHandler) RemoveTaskTag(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	tag := c.Param("tag")
	if tag == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Tag parameter is required"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.RemoveTag(tag) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Tag not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// AddTaskReminder adds a reminder to a task
func (h *TaskHandler) AddTaskReminder(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var reminderData struct {
		Time        string `json:"time" binding:"required"`
		Description string `json:"description"`
		Method      string `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reminderData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	reminderTime, err := time.Parse(time.RFC3339, reminderData.Time)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid time format, use RFC3339"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task.AddReminder(reminderTime, reminderData.Description, reminderData.Method)

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// RemoveTaskReminder removes a reminder from a task
func (h *TaskHandler) RemoveTaskReminder(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	reminderID, err := strconv.Atoi(c.Param("reminderId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.RemoveReminder(reminderID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// GetTaskReminders gets all reminders for a task
func (h *TaskHandler) GetTaskReminders(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"reminders": task.Reminders})
}

// AddTaskSubtask adds a subtask to a task
func (h *TaskHandler) AddTaskSubtask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var subtaskData struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&subtaskData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	task.AddSubtask(subtaskData.Title)

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// CompleteTaskSubtask marks a subtask as completed
func (h *TaskHandler) CompleteTaskSubtask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	subtaskID, err := strconv.Atoi(c.Param("subtaskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.CompleteSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// ReopenTaskSubtask marks a subtask as not completed
func (h *TaskHandler) ReopenTaskSubtask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	subtaskID, err := strconv.Atoi(c.Param("subtaskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.ReopenSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// DeleteTaskSubtask removes a subtask from a task
func (h *TaskHandler) DeleteTaskSubtask(c *gin.Context) {
	taskID, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	subtaskID, err := strconv.Atoi(c.Param("subtaskId"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID"})
		return
	}

	task, err := h.storage.Get(taskID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if task == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	if !task.DeleteSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"task": task})
}

// BatchCreateTasks creates multiple tasks at once
func (h *TaskHandler) BatchCreateTasks(c *gin.Context) {
	var tasksData struct {
		Tasks []models.Task `json:"tasks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tasksData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tasks := make([]*models.Task, len(tasksData.Tasks))
	for i := range tasksData.Tasks {
		tasks[i] = &tasksData.Tasks[i]
	}

	if err := h.storage.BatchCreate(tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"tasks": tasks})
}

// BatchUpdateTasks updates multiple tasks at once
func (h *TaskHandler) BatchUpdateTasks(c *gin.Context) {
	var tasksData struct {
		Tasks []models.Task `json:"tasks" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tasksData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	tasks := make([]*models.Task, len(tasksData.Tasks))
	for i := range tasksData.Tasks {
		tasks[i] = &tasksData.Tasks[i]
	}

	if err := h.storage.BatchUpdate(tasks); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// BatchDeleteTasks deletes multiple tasks at once
func (h *TaskHandler) BatchDeleteTasks(c *gin.Context) {
	var tasksData struct {
		TaskIds []int `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tasksData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.BatchDelete(tasksData.TaskIds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// BatchCompleteTasks marks multiple tasks as completed
func (h *TaskHandler) BatchCompleteTasks(c *gin.Context) {
	var tasksData struct {
		TaskIds []int `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tasksData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.BatchComplete(tasksData.TaskIds); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}
