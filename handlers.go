package main

import (
	"net/http"
	"sort"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

// TaskHandler handles HTTP requests for tasks
type TaskHandler struct {
	storage Storage
}

// NewTaskHandler creates a new task handler
func NewTaskHandler(storage Storage) *TaskHandler {
	return &TaskHandler{
		storage: storage,
	}
}

// GetTasks returns all tasks with optional filtering and sorting
func (h *TaskHandler) GetTasks(c *gin.Context) {
	// Get all tasks first
	tasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Parse filter options
	var filterOptions FilterOptions

	// Filter by completion status
	if completedStr := c.Query("completed"); completedStr != "" {
		completed := completedStr == "true"
		filterOptions.Completed = &completed
	}

	// Filter by category
	if category := c.Query("category"); category != "" {
		filterOptions.Category = category
	}

	// Filter by priority
	if priority := c.Query("priority"); priority != "" {
		filterOptions.Priority = Priority(priority)
	}

	// Filter by tag
	if tag := c.Query("tag"); tag != "" {
		filterOptions.Tag = tag
	}

	// Filter by due date before
	if dueBefore := c.Query("dueBefore"); dueBefore != "" {
		if t, err := time.Parse(time.RFC3339, dueBefore); err == nil {
			filterOptions.DueBefore = t
		}
	}

	// Filter by due date after
	if dueAfter := c.Query("dueAfter"); dueAfter != "" {
		if t, err := time.Parse(time.RFC3339, dueAfter); err == nil {
			filterOptions.DueAfter = t
		}
	}

	// Filter by creation date after
	if createdAfter := c.Query("createdAfter"); createdAfter != "" {
		if t, err := time.Parse(time.RFC3339, createdAfter); err == nil {
			filterOptions.CreatedAfter = t
		}
	}

	// Filter by assignee
	if assignedTo := c.Query("assignedTo"); assignedTo != "" {
		filterOptions.AssignedTo = assignedTo
	}

	// Filter by creator
	if createdBy := c.Query("createdBy"); createdBy != "" {
		filterOptions.CreatedBy = createdBy
	}

	// Apply filters if any were specified
	if filterOptions != (FilterOptions{}) {
		filteredTasks, err := h.storage.Filter(filterOptions)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		tasks = filteredTasks
	}

	// Parse sort options
	var sortOptions SortOptions

	// Sort by field
	if field := c.Query("sortBy"); field != "" {
		sortOptions.Field = field
	} else {
		sortOptions.Field = "createdAt" // Default sort field
	}

	// Sort direction
	if direction := c.Query("sortDir"); direction != "" {
		sortOptions.Direction = direction
	} else {
		sortOptions.Direction = "desc" // Default sort direction
	}

	// Apply sorting
	sortedTasks := h.storage.Sort(tasks, sortOptions)

	// Pagination
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	perPage, _ := strconv.Atoi(c.DefaultQuery("perPage", "20"))

	// Validate pagination parameters
	if page < 1 {
		page = 1
	}
	if perPage < 1 || perPage > 100 {
		perPage = 20
	}

	// Calculate start and end indices for pagination
	startIndex := (page - 1) * perPage
	endIndex := startIndex + perPage

	// Check if start index is valid
	if startIndex >= len(sortedTasks) {
		// Return empty array if page is out of range
		c.JSON(http.StatusOK, gin.H{
			"tasks": []interface{}{},
			"pagination": gin.H{
				"total":       len(sortedTasks),
				"currentPage": page,
				"perPage":     perPage,
				"totalPages":  (len(sortedTasks) + perPage - 1) / perPage,
				"hasMore":     false,
			},
		})
		return
	}

	// Adjust end index if it exceeds array length
	if endIndex > len(sortedTasks) {
		endIndex = len(sortedTasks)
	}

	// Extract the page of tasks
	pagedTasks := sortedTasks[startIndex:endIndex]

	// Construct the response with pagination info
	c.JSON(http.StatusOK, gin.H{
		"tasks": pagedTasks,
		"pagination": gin.H{
			"total":       len(sortedTasks),
			"currentPage": page,
			"perPage":     perPage,
			"totalPages":  (len(sortedTasks) + perPage - 1) / perPage,
			"hasMore":     endIndex < len(sortedTasks),
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

	c.JSON(http.StatusOK, task)
}

// CreateTask creates a new task
func (h *TaskHandler) CreateTask(c *gin.Context) {
	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.storage.Add(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, task)
}

// UpdateTask updates an existing task
func (h *TaskHandler) UpdateTask(c *gin.Context) {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID"})
		return
	}

	var task Task
	if err := c.ShouldBindJSON(&task); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.ID = id
	if err := h.storage.Update(&task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
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

	c.Status(http.StatusNoContent)
}

// HealthCheck responds with server health information
func (h *TaskHandler) HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"status": "ok",
	})
}

// GetCategories returns all unique categories
func (h *TaskHandler) GetCategories(c *gin.Context) {
	tasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"categories": uniqueCategories,
	})
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

// AddTaskNote adds a note to a task
func (h *TaskHandler) AddTaskNote(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var noteInput struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&noteInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get username from context (set by AuthMiddleware)
	username, exists := c.Get("username")
	if !exists {
		// If username doesn't exist in context, use "unknown" as fallback
		task.AddNote(noteInput.Content, "unknown")
	} else {
		task.AddNote(noteInput.Content, username.(string))
	}
	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// UpdateTaskNote updates a note on a task
func (h *TaskHandler) UpdateTaskNote(c *gin.Context) {
	idStr := c.Param("id")
	noteIDStr := c.Param("noteId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID format"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var noteInput struct {
		Content string `json:"content" binding:"required"`
	}

	if err := c.ShouldBindJSON(&noteInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !task.UpdateNote(noteID, noteInput.Content) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Note not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTaskNote deletes a note from a task
func (h *TaskHandler) DeleteTaskNote(c *gin.Context) {
	idStr := c.Param("id")
	noteIDStr := c.Param("noteId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	noteID, err := strconv.Atoi(noteIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid note ID format"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
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

	c.JSON(http.StatusOK, task)
}

// AddTaskTag adds a tag to a task
func (h *TaskHandler) AddTaskTag(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Task not found"})
		return
	}

	var tagInput struct {
		Tag string `json:"tag" binding:"required"`
	}

	if err := c.ShouldBindJSON(&tagInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.AddTag(tagInput.Tag)
	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// RemoveTaskTag removes a tag from a task
func (h *TaskHandler) RemoveTaskTag(c *gin.Context) {
	idStr := c.Param("id")
	tag := c.Param("tag")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid ID format"})
		return
	}

	task, err := h.storage.Get(id)
	if err != nil {
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

	c.JSON(http.StatusOK, task)
}

// GetTags returns all unique tags
func (h *TaskHandler) GetTags(c *gin.Context) {
	tasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
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

	c.JSON(http.StatusOK, gin.H{
		"tags": uniqueTags,
	})
}

// AddTaskReminder adds a reminder to a task
func (h *TaskHandler) AddTaskReminder(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
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

	var reminderInput struct {
		Time        time.Time `json:"time" binding:"required"`
		Description string    `json:"description"`
		Method      string    `json:"method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&reminderInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.AddReminder(reminderInput.Time, reminderInput.Description, reminderInput.Method)
	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// RemoveTaskReminder removes a reminder from a task
func (h *TaskHandler) RemoveTaskReminder(c *gin.Context) {
	idStr := c.Param("id")
	reminderIDStr := c.Param("reminderId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	reminderID, err := strconv.Atoi(reminderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid reminder ID format"})
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

	if !task.RemoveReminder(reminderID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Reminder not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskReminders gets all reminders for a task
func (h *TaskHandler) GetTaskReminders(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
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

	c.JSON(http.StatusOK, gin.H{"reminders": task.Reminders})
}

// AddTaskSubtask adds a subtask to a task
func (h *TaskHandler) AddTaskSubtask(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
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

	var subtaskInput struct {
		Title string `json:"title" binding:"required"`
	}

	if err := c.ShouldBindJSON(&subtaskInput); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	task.AddSubtask(subtaskInput.Title)
	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// CompleteTaskSubtask marks a subtask as completed
func (h *TaskHandler) CompleteTaskSubtask(c *gin.Context) {
	idStr := c.Param("id")
	subtaskIDStr := c.Param("subtaskId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	subtaskID, err := strconv.Atoi(subtaskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID format"})
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

	if !task.CompleteSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// ReopenTaskSubtask marks a subtask as not completed
func (h *TaskHandler) ReopenTaskSubtask(c *gin.Context) {
	idStr := c.Param("id")
	subtaskIDStr := c.Param("subtaskId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	subtaskID, err := strconv.Atoi(subtaskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID format"})
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

	if !task.ReopenSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// DeleteTaskSubtask deletes a subtask
func (h *TaskHandler) DeleteTaskSubtask(c *gin.Context) {
	idStr := c.Param("id")
	subtaskIDStr := c.Param("subtaskId")

	id, err := strconv.Atoi(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid task ID format"})
		return
	}

	subtaskID, err := strconv.Atoi(subtaskIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid subtask ID format"})
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

	if !task.DeleteSubtask(subtaskID) {
		c.JSON(http.StatusNotFound, gin.H{"error": "Subtask not found"})
		return
	}

	if err := h.storage.Update(task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, task)
}

// GetTaskAnalytics returns detailed analytics about tasks
func (h *TaskHandler) GetTaskAnalytics(c *gin.Context) {
	tasks, err := h.storage.GetAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Group tasks by category
	categoryCounts := make(map[string]int)
	for _, task := range tasks {
		categoryCounts[task.Category]++
	}

	// Calculate completion rates
	completedTasks := 0
	for _, task := range tasks {
		if task.Completed {
			completedTasks++
		}
	}

	completionRate := 0.0
	if len(tasks) > 0 {
		completionRate = float64(completedTasks) / float64(len(tasks)) * 100
	}

	// Calculate priority distribution
	priorityCounts := map[string]int{
		"low":    0,
		"medium": 0,
		"high":   0,
	}

	for _, task := range tasks {
		priorityCounts[string(task.Priority)]++
	}

	// Calculate completion time metrics
	var completionTimes []float64
	for _, task := range tasks {
		if task.Completed && !task.CompletedAt.IsZero() && !task.CreatedAt.IsZero() {
			completionTime := task.CompletedAt.Sub(task.CreatedAt).Hours()
			completionTimes = append(completionTimes, completionTime)
		}
	}

	averageCompletionTime := 0.0
	if len(completionTimes) > 0 {
		total := 0.0
		for _, t := range completionTimes {
			total += t
		}
		averageCompletionTime = total / float64(len(completionTimes))
	}

	// Calculate tag usage
	tagCounts := make(map[string]int)
	for _, task := range tasks {
		for _, tag := range task.Tags {
			tagCounts[tag]++
		}
	}

	// Find most used tags
	type TagCount struct {
		Tag   string `json:"tag"`
		Count int    `json:"count"`
	}

	tagCountList := make([]TagCount, 0, len(tagCounts))
	for tag, count := range tagCounts {
		tagCountList = append(tagCountList, TagCount{Tag: tag, Count: count})
	}

	// Sort by count (descending)
	sort.Slice(tagCountList, func(i, j int) bool {
		return tagCountList[i].Count > tagCountList[j].Count
	})

	// Get top 10 tags
	topTags := tagCountList
	if len(topTags) > 10 {
		topTags = topTags[:10]
	}

	// Calculate overdue tasks
	overdueTasks := 0
	for _, task := range tasks {
		if !task.Completed && !task.DueDate.IsZero() && task.DueDate.Before(time.Now()) {
			overdueTasks++
		}
	}

	// Weekly activity analysis
	now := time.Now()
	startOfWeek := now.AddDate(0, 0, -int(now.Weekday()))
	weeklyActivity := make(map[string]int)

	// Initialize days of the week
	for i := 0; i < 7; i++ {
		day := startOfWeek.AddDate(0, 0, i).Format("Monday")
		weeklyActivity[day] = 0
	}

	// Count tasks created/completed in each day of the current week
	for _, task := range tasks {
		// Count task creations
		if task.CreatedAt.After(startOfWeek) && task.CreatedAt.Before(startOfWeek.AddDate(0, 0, 7)) {
			day := task.CreatedAt.Format("Monday")
			weeklyActivity[day]++
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"total_tasks":           len(tasks),
		"completed_tasks":       completedTasks,
		"completion_rate":       completionRate,
		"category_distribution": categoryCounts,
		"priority_distribution": priorityCounts,
		"avg_completion_time":   averageCompletionTime,
		"top_tags":              topTags,
		"overdue_tasks":         overdueTasks,
		"weekly_activity":       weeklyActivity,
	})
}

// BatchCreateTasks creates multiple tasks at once
func (h *TaskHandler) BatchCreateTasks(c *gin.Context) {
	var tasks []*Task
	if err := c.ShouldBindJSON(&tasks); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(tasks) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tasks provided"})
		return
	}

	if len(tasks) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many tasks. Maximum allowed is 100"})
		return
	}

	createdTasks := make([]*Task, 0, len(tasks))
	for _, task := range tasks {
		if err := h.storage.Add(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		createdTasks = append(createdTasks, task)
	}

	c.JSON(http.StatusCreated, gin.H{"tasks": createdTasks})
}

// BatchUpdateTasks updates multiple tasks at once
func (h *TaskHandler) BatchUpdateTasks(c *gin.Context) {
	var tasksToUpdate []*Task
	if err := c.ShouldBindJSON(&tasksToUpdate); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(tasksToUpdate) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No tasks provided"})
		return
	}

	if len(tasksToUpdate) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many tasks. Maximum allowed is 100"})
		return
	}

	updatedTasks := make([]*Task, 0, len(tasksToUpdate))
	for _, task := range tasksToUpdate {
		if task.ID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required for updates"})
			return
		}

		// Verify task exists
		existingTask, err := h.storage.Get(task.ID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if existingTask == nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "Task not found: " + strconv.Itoa(task.ID)})
			return
		}

		if err := h.storage.Update(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		updatedTasks = append(updatedTasks, task)
	}

	c.JSON(http.StatusOK, gin.H{"tasks": updatedTasks})
}

// BatchDeleteTasks deletes multiple tasks at once
func (h *TaskHandler) BatchDeleteTasks(c *gin.Context) {
	var input struct {
		TaskIDs []int `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.TaskIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No task IDs provided"})
		return
	}

	if len(input.TaskIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many tasks. Maximum allowed is 100"})
		return
	}

	deletedIDs := make([]int, 0, len(input.TaskIDs))
	for _, id := range input.TaskIDs {
		if err := h.storage.Delete(id); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		deletedIDs = append(deletedIDs, id)
	}

	c.JSON(http.StatusOK, gin.H{"deletedTaskIds": deletedIDs})
}

// BatchCompleteTasks marks multiple tasks as completed at once
func (h *TaskHandler) BatchCompleteTasks(c *gin.Context) {
	var input struct {
		TaskIDs []int `json:"taskIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if len(input.TaskIDs) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No task IDs provided"})
		return
	}

	if len(input.TaskIDs) > 100 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Too many tasks. Maximum allowed is 100"})
		return
	}

	completedTasks := make([]*Task, 0, len(input.TaskIDs))
	for _, id := range input.TaskIDs {
		task, err := h.storage.Get(id)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if task == nil {
			continue // Skip non-existent tasks
		}

		task.MarkCompleted()
		if err := h.storage.Update(task); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		completedTasks = append(completedTasks, task)
	}

	c.JSON(http.StatusOK, gin.H{"tasks": completedTasks})
}
