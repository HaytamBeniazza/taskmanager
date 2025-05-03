package handlers

import (
	"time"

	"github.com/HaytamBeniazza/taskmanager/internal/middleware"
	"github.com/gin-gonic/gin"
)

// APIVersion is the current version of the API
const APIVersion = "1.0.0"

// SetupRouter configures the Gin router with all routes and middleware
func SetupRouter(taskHandler *TaskHandler) *gin.Engine {
	// Create a new Gin router
	r := gin.New()

	// Apply global middleware
	r.Use(gin.Recovery())
	r.Use(middleware.Logger())
	r.Use(middleware.RequestID())
	r.Use(middleware.CORSMiddleware())
	r.Use(middleware.APIVersioning(APIVersion))
	r.Use(middleware.RequestSizeLimiter(10 << 20)) // 10 MB max request size

	// Basic health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":    "up",
			"time":      time.Now().Format(time.RFC3339),
			"version":   APIVersion,
			"serverID":  "task-manager-01",
			"buildInfo": "1.0.0-8675309",
		})
	})

	// API routes - versioned under /api/v1
	api := r.Group("/api/v1")
	{
		// Tasks endpoints
		tasks := api.Group("/tasks")
		{
			tasks.GET("", taskHandler.GetTasks)
			tasks.GET("/search", taskHandler.SearchTasks)
			tasks.GET("/:id", taskHandler.GetTask)
			tasks.POST("", taskHandler.CreateTask)
			tasks.PUT("/:id", taskHandler.UpdateTask)
			tasks.DELETE("/:id", taskHandler.DeleteTask)

			// Task notes
			tasks.POST("/:id/notes", taskHandler.AddTaskNote)
			tasks.PUT("/:id/notes/:noteId", taskHandler.UpdateTaskNote)
			tasks.DELETE("/:id/notes/:noteId", taskHandler.DeleteTaskNote)

			// Task tags
			tasks.POST("/:id/tags", taskHandler.AddTaskTag)
			tasks.DELETE("/:id/tags/:tag", taskHandler.RemoveTaskTag)

			// Task reminders
			tasks.POST("/:id/reminders", taskHandler.AddTaskReminder)
			tasks.DELETE("/:id/reminders/:reminderId", taskHandler.RemoveTaskReminder)
			tasks.GET("/:id/reminders", taskHandler.GetTaskReminders)

			// Task subtasks
			tasks.POST("/:id/subtasks", taskHandler.AddTaskSubtask)
			tasks.PUT("/:id/subtasks/:subtaskId/complete", taskHandler.CompleteTaskSubtask)
			tasks.PUT("/:id/subtasks/:subtaskId/reopen", taskHandler.ReopenTaskSubtask)
			tasks.DELETE("/:id/subtasks/:subtaskId", taskHandler.DeleteTaskSubtask)

			// Batch operations
			tasks.POST("/batch/create", taskHandler.BatchCreateTasks)
			tasks.PUT("/batch/update", taskHandler.BatchUpdateTasks)
			tasks.DELETE("/batch/delete", taskHandler.BatchDeleteTasks)
			tasks.PUT("/batch/complete", taskHandler.BatchCompleteTasks)
		}

		// Metadata endpoints
		api.GET("/categories", taskHandler.GetCategories)
		api.GET("/tags", taskHandler.GetTags)

		// Statistics endpoints
		api.GET("/stats", taskHandler.GetTaskStats)
		api.GET("/analytics", taskHandler.GetTaskAnalytics)
	}

	return r
}
