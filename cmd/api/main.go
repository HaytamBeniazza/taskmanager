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
		log.Println("Initializing database with sample data...")
		if err := createInitialData(taskStorage); err != nil {
			log.Fatalf("Failed to create sample data: %v", err)
		}
	}

	// Create task handler
	taskHandler := handlers.NewTaskHandler(taskStorage)

	// Set up router with all routes
	router := handlers.SetupRouter(taskHandler)

	// Start server
	addr := fmt.Sprintf(":%d", cfg.Port)
	log.Printf("Server starting on %s in %s mode", addr, cfg.Env)
	if err := router.Run(addr); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// createInitialData adds some sample tasks to the database
func createInitialData(storage storage.TaskStorage) error {
	now := time.Now()
	tomorrow := now.Add(24 * time.Hour)
	nextWeek := now.Add(7 * 24 * time.Hour)

	// Sample tasks
	tasks := []*models.Task{
		{
			Title:       "Complete project proposal",
			Description: "Draft and submit the initial project proposal document",
			Priority:    models.High,
			Category:    "Work",
			DueDate:     tomorrow,
			Tags:        []string{"work", "proposal", "urgent"},
			CreatedAt:   now,
			Subtasks: []models.Subtask{
				{ID: 1, Title: "Research competitors", Completed: true},
				{ID: 2, Title: "Create outline", Completed: true},
				{ID: 3, Title: "Write first draft", Completed: false},
				{ID: 4, Title: "Get feedback", Completed: false},
				{ID: 5, Title: "Finalize document", Completed: false},
			},
			CreatedBy: "admin",
		},
		{
			Title:       "Buy groceries",
			Description: "Get weekly groceries from the supermarket",
			Priority:    models.Medium,
			Category:    "Personal",
			DueDate:     tomorrow,
			Tags:        []string{"shopping", "food", "weekly"},
			CreatedAt:   now,
			Subtasks: []models.Subtask{
				{ID: 1, Title: "Make shopping list", Completed: true},
				{ID: 2, Title: "Check fridge for needed items", Completed: false},
			},
			CreatedBy: "admin",
		},
		{
			Title:       "Schedule dentist appointment",
			Description: "Call the dentist for a regular checkup",
			Priority:    models.Low,
			Category:    "Health",
			DueDate:     nextWeek,
			Tags:        []string{"health", "appointment"},
			CreatedAt:   now,
			CreatedBy:   "admin",
		},
		{
			Title:       "Review quarterly budget",
			Description: "Review and adjust the quarterly budget plan",
			Priority:    models.High,
			Category:    "Finance",
			DueDate:     nextWeek,
			Tags:        []string{"finance", "quarterly", "budget"},
			CreatedAt:   now.Add(-48 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Plan team building event",
			Description: "Organize a team building activity for the department",
			Priority:    models.Medium,
			Category:    "Work",
			DueDate:     nextWeek.Add(72 * time.Hour),
			Tags:        []string{"work", "team", "event"},
			CreatedAt:   now.Add(-24 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Renew car insurance",
			Description: "Shop around for the best car insurance rates",
			Priority:    models.High,
			Category:    "Finance",
			DueDate:     nextWeek.Add(48 * time.Hour),
			Tags:        []string{"finance", "car", "insurance"},
			CreatedAt:   now.Add(-72 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Prepare for presentation",
			Description: "Create slides and prepare talking points for the client presentation",
			Priority:    models.High,
			Category:    "Work",
			DueDate:     tomorrow.Add(24 * time.Hour),
			Tags:        []string{"work", "presentation", "client"},
			CreatedAt:   now.Add(-12 * time.Hour),
			Completed:   true,
			CompletedAt: now.Add(-1 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Fix leaking faucet",
			Description: "Replace the washer in the kitchen sink faucet",
			Priority:    models.Medium,
			Category:    "Home",
			DueDate:     tomorrow.Add(48 * time.Hour),
			Tags:        []string{"home", "repair", "plumbing"},
			CreatedAt:   now.Add(-36 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Call mom",
			Description: "Weekly call with mom to catch up",
			Priority:    models.Medium,
			Category:    "Personal",
			DueDate:     now.Add(3 * 24 * time.Hour),
			Tags:        []string{"personal", "family", "weekly"},
			CreatedAt:   now.Add(-2 * time.Hour),
			CreatedBy:   "admin",
		},
		{
			Title:       "Book vacation flights",
			Description: "Search for and book flights for summer vacation",
			Priority:    models.Medium,
			Category:    "Travel",
			DueDate:     now.Add(30 * 24 * time.Hour),
			Tags:        []string{"travel", "vacation", "planning"},
			CreatedAt:   now.Add(-5 * 24 * time.Hour),
			CreatedBy:   "admin",
		},
	}

	// Add tasks to database
	for _, task := range tasks {
		if err := storage.Add(task); err != nil {
			return err
		}
	}

	return nil
}
