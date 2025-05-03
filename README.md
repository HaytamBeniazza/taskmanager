# Go Task Manager

A modern task management application built with Go backend and React frontend. This project demonstrates a full-stack application with persistent storage, RESTful API, and a responsive UI.

## Recent Updates

- **Enhanced UI Theme**: Implemented a modern theme with improved color scheme, typography, and component styling
- **Dashboard Improvements**: Added better visualization of task statistics, priority distribution, and category breakdown
- **Advanced Task Filtering**: New filtering options including due date ranges, priority filtering, and tag-based search
- **Task List Enhancements**: Added grid and list view options with improved sorting capabilities
- **UI Components**: Redesigned cards, buttons, and chips for better user experience
- **Frontend Dependencies**: Updated and optimized frontend dependencies
- **Project Structure**: Reorganized codebase to follow professional Go project architecture patterns

## Features

### Backend (Go):
- RESTful API using Gin framework
- Persistent storage with BoltDB (a pure Go key-value database)
- Task management with CRUD operations
- Task categorization, priorities, and tagging
- Task notes and search functionality
- Statistics and reporting

### Frontend (React):
- Modern UI with Material-UI components
- Responsive design
- Dashboard with task statistics
- Task listing with filtering and search
- Task creation and editing
- Form validation

## Project Structure

```
taskmanager/
├── cmd/                  # Application entry points
│   └── api/              # API server executable
│       └── main.go       # Main application entry point
├── config/               # Configuration
│   └── config.go         # Application configuration
├── internal/             # Private application code
│   ├── handlers/         # HTTP request handlers
│   ├── middleware/       # HTTP middleware components
│   ├── models/           # Data models
│   ├── storage/          # Database interactions
│   └── utils/            # Utility functions
├── pkg/                  # Public libraries for external usage
├── frontend/             # React frontend application
└── README.md             # Project documentation
```

## Tech Stack

### Backend:
- Go (Golang)
- Gin Web Framework
- BoltDB for storage

### Frontend:
- React
- TypeScript
- Material-UI
- React Router
- Axios for API communication

## Getting Started

### Prerequisites
- Go 1.16 or higher
- Node.js 14.x or higher
- npm 6.x or higher

### Backend Setup
1. Clone the repository
2. Navigate to the project root
3. Run the Go application:
   ```
   cd cmd/api
   go run main.go
   ```
   
### Frontend Setup
1. Navigate to the frontend directory:
   ```
   cd frontend
   ```
2. Install dependencies:
   ```
   npm install
   ```
3. Start the development server:
   ```
   npm start
   ```
4. Open http://localhost:3000 in your browser

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | /health | Health check endpoint |
| GET    | /stats | Get task statistics |
| GET    | /analytics | Get task analytics data |
| GET    | /tasks | Get all tasks |
| GET    | /tasks/search?q={query} | Search tasks |
| GET    | /tasks/categories | Get all task categories |
| GET    | /tasks/tags | Get all task tags |
| GET    | /tasks/:id | Get a specific task |
| POST   | /tasks | Create a new task |
| PUT    | /tasks/:id | Update a task |
| DELETE | /tasks/:id | Delete a task |
| POST   | /tasks/batch/complete | Complete multiple tasks |
| DELETE | /tasks/batch/delete | Delete multiple tasks |
| POST   | /tasks/:id/notes | Add a note to a task |
| PUT    | /tasks/:id/notes/:noteId | Update a task note |
| DELETE | /tasks/:id/notes/:noteId | Delete a task note |
| POST   | /tasks/:id/tags | Add a tag to a task |
| DELETE | /tasks/:id/tags/:tag | Remove a tag from a task |

## Configuration

The backend can be configured using environment variables:

- `PORT`: HTTP port (default: 8080)
- `ENV`: Environment (development/production)
- `CORS_ORIGIN`: CORS allowed origin (default: *)
- `DB_PATH`: Path to the database file (default: taskmanager.db)

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with ❤️ using Go and React
- Material-UI for the beautiful UI components
- BoltDB for the pure Go database solution 