package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/HaytamBeniazza/taskmanager/config"
	"github.com/HaytamBeniazza/taskmanager/internal/handlers"
	"github.com/HaytamBeniazza/taskmanager/internal/models"
	"github.com/HaytamBeniazza/taskmanager/internal/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	// Load configuration
	cfg := config.LoadConfig()

	// Create BoltDB storage
	taskStorage, err := storage.NewBoltDBStorage(cfg.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer taskStorage.Close()

	// Check if we need to create sample data
	createSampleData := false
	if _, err := os.Stat(cfg.DBPath); os.IsNotExist(err) {
		createSampleData = true
	}

	// Create sample tasks if this is a new database
	if createSampleData {
		log.Println("Creating sample tasks...")

		task1 := models.NewTask("Learn Go basics", "Study syntax, types, and basic concepts")
		task1.Category = "Learning"
		task1.Priority = models.Medium
		task1.DueDate = time.Now().Add(24 * time.Hour)
		task1.AddTag("go")
		task1.AddTag("programming")
		task1.AddNote("Focus on understanding goroutines and channels", "system")

		task2 := models.NewTask("Build a web API", "Create a task manager REST API in Go")
		task2.Category = "Development"
		task2.Priority = models.High
		task2.DueDate = time.Now().Add(72 * time.Hour)
		task2.AddTag("go")
		task2.AddTag("api")
		task2.AddTag("rest")
		task2.AddNote("Use Gin framework for routing", "system")

		task3 := models.NewTask("Learn concurrency", "Study goroutines and channels")
		task3.Category = "Learning"
		task3.Priority = models.Low
		task3.AddTag("go")
		task3.AddTag("concurrency")

		// Add sample tasks
		if err := taskStorage.Add(task1); err != nil {
			log.Printf("Failed to add sample task 1: %v", err)
		}
		if err := taskStorage.Add(task2); err != nil {
			log.Printf("Failed to add sample task 2: %v", err)
		}
		if err := taskStorage.Add(task3); err != nil {
			log.Printf("Failed to add sample task 3: %v", err)
		}
	}

	// Create task handler
	taskHandler := handlers.NewTaskHandler(taskStorage)

	// Set up Gin router
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{cfg.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	router.GET("/health", taskHandler.HealthCheck)
	router.GET("/stats", taskHandler.GetTaskStats)
	router.GET("/analytics", taskHandler.GetTaskAnalytics)

	// Tasks endpoints
	router.GET("/tasks", taskHandler.GetTasks)
	router.GET("/tasks/search", taskHandler.SearchTasks)
	router.GET("/tasks/categories", taskHandler.GetCategories)
	router.GET("/tasks/tags", taskHandler.GetTags)
	router.GET("/tasks/:id", taskHandler.GetTask)
	router.POST("/tasks", taskHandler.CreateTask)
	router.PUT("/tasks/:id", taskHandler.UpdateTask)
	router.DELETE("/tasks/:id", taskHandler.DeleteTask)

	// Batch operations
	router.POST("/tasks/batch/create", taskHandler.BatchCreateTasks)
	router.PUT("/tasks/batch/update", taskHandler.BatchUpdateTasks)
	router.DELETE("/tasks/batch/delete", taskHandler.BatchDeleteTasks)
	router.POST("/tasks/batch/complete", taskHandler.BatchCompleteTasks)

	// Task notes endpoints
	router.POST("/tasks/:id/notes", taskHandler.AddTaskNote)
	router.PUT("/tasks/:id/notes/:noteId", taskHandler.UpdateTaskNote)
	router.DELETE("/tasks/:id/notes/:noteId", taskHandler.DeleteTaskNote)

	// Task tags endpoints
	router.POST("/tasks/:id/tags", taskHandler.AddTaskTag)
	router.DELETE("/tasks/:id/tags/:tag", taskHandler.RemoveTaskTag)

	// Task reminder endpoints
	router.GET("/tasks/:id/reminders", taskHandler.GetTaskReminders)
	router.POST("/tasks/:id/reminders", taskHandler.AddTaskReminder)
	router.DELETE("/tasks/:id/reminders/:reminderId", taskHandler.RemoveTaskReminder)

	// Task subtask endpoints
	router.POST("/tasks/:id/subtasks", taskHandler.AddTaskSubtask)
	router.PUT("/tasks/:id/subtasks/:subtaskId/complete", taskHandler.CompleteTaskSubtask)
	router.PUT("/tasks/:id/subtasks/:subtaskId/reopen", taskHandler.ReopenTaskSubtask)
	router.DELETE("/tasks/:id/subtasks/:subtaskId", taskHandler.DeleteTaskSubtask)

	// Start the server
	port := fmt.Sprintf(":%d", cfg.Port)
	fmt.Printf("Server starting on port %d in %s mode\n", cfg.Port, cfg.Env)
	fmt.Printf("Using database: %s\n", cfg.DBPath)
	router.Run(port)
}
