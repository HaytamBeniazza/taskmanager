package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Config holds the application configuration
type Config struct {
	Port       int
	Env        string
	CorsOrigin string
	DBPath     string
}

// LoadConfig loads configuration from environment variables
func LoadConfig() Config {
	// Default values
	config := Config{
		Port:       8080,
		Env:        "development",
		CorsOrigin: "*",
		DBPath:     "taskmanager.db",
	}

	// Override from environment variables if provided
	if port, err := strconv.Atoi(os.Getenv("PORT")); err == nil && port > 0 {
		config.Port = port
	}

	if env := os.Getenv("ENV"); env != "" {
		config.Env = env
	}

	if corsOrigin := os.Getenv("CORS_ORIGIN"); corsOrigin != "" {
		config.CorsOrigin = corsOrigin
	}

	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		config.DBPath = dbPath
	}

	return config
}

func main() {
	// Load configuration
	config := LoadConfig()

	// Create BoltDB storage
	storage, err := NewBoltDBStorage(config.DBPath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer storage.Close()

	// Check if we need to create sample data
	createSampleData := false
	if _, err := os.Stat(config.DBPath); os.IsNotExist(err) {
		createSampleData = true
	}

	// Create sample tasks if this is a new database
	if createSampleData {
		log.Println("Creating sample tasks...")

		task1 := NewTask("Learn Go basics", "Study syntax, types, and basic concepts")
		task1.Category = "Learning"
		task1.Priority = Medium
		task1.DueDate = time.Now().Add(24 * time.Hour)
		task1.AddTag("go")
		task1.AddTag("programming")
		task1.AddNote("Focus on understanding goroutines and channels")

		task2 := NewTask("Build a web API", "Create a task manager REST API in Go")
		task2.Category = "Development"
		task2.Priority = High
		task2.DueDate = time.Now().Add(72 * time.Hour)
		task2.AddTag("go")
		task2.AddTag("api")
		task2.AddTag("rest")
		task2.AddNote("Use Gin framework for routing")

		task3 := NewTask("Learn concurrency", "Study goroutines and channels")
		task3.Category = "Learning"
		task3.Priority = Low
		task3.AddTag("go")
		task3.AddTag("concurrency")

		// Add sample tasks
		if err := storage.Add(task1); err != nil {
			log.Printf("Failed to add sample task 1: %v", err)
		}
		if err := storage.Add(task2); err != nil {
			log.Printf("Failed to add sample task 2: %v", err)
		}
		if err := storage.Add(task3); err != nil {
			log.Printf("Failed to add sample task 3: %v", err)
		}
	}

	// Create task handler
	taskHandler := NewTaskHandler(storage)

	// Set up Gin router
	if config.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.Default()

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{config.CorsOrigin},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// API routes
	router.GET("/health", taskHandler.HealthCheck)
	router.GET("/stats", taskHandler.GetTaskStats)

	// Tasks endpoints
	router.GET("/tasks", taskHandler.GetTasks)
	router.GET("/tasks/search", taskHandler.SearchTasks)
	router.GET("/tasks/categories", taskHandler.GetCategories)
	router.GET("/tasks/tags", taskHandler.GetTags)
	router.GET("/tasks/:id", taskHandler.GetTask)
	router.POST("/tasks", taskHandler.CreateTask)
	router.PUT("/tasks/:id", taskHandler.UpdateTask)
	router.DELETE("/tasks/:id", taskHandler.DeleteTask)

	// Task notes endpoints
	router.POST("/tasks/:id/notes", taskHandler.AddTaskNote)
	router.PUT("/tasks/:id/notes/:noteId", taskHandler.UpdateTaskNote)
	router.DELETE("/tasks/:id/notes/:noteId", taskHandler.DeleteTaskNote)

	// Task tags endpoints
	router.POST("/tasks/:id/tags", taskHandler.AddTaskTag)
	router.DELETE("/tasks/:id/tags/:tag", taskHandler.RemoveTaskTag)

	// Start the server
	port := fmt.Sprintf(":%d", config.Port)
	fmt.Printf("Server starting on port %d in %s mode\n", config.Port, config.Env)
	fmt.Printf("Using database: %s\n", config.DBPath)
	router.Run(port)
}
