# Go Task Manager

A modern task management application built with Go backend and React frontend. This project demonstrates a full-stack application with persistent storage, RESTful API, and a responsive UI.

## Recent Updates

- **Clean Architecture**: Completely restructured the codebase to follow Go's clean architecture principles
- **Versioned API**: API routes are now versioned under `/api/v1`
- **Middleware Framework**: Added comprehensive middleware support for logging, CORS, rate limiting
- **Enhanced UI Theme**: Implemented a modern theme with improved color scheme, typography, and component styling
- **Dashboard Improvements**: Added better visualization of task statistics, priority distribution, and category breakdown
- **Advanced Task Filtering**: New filtering options including due date ranges, priority filtering, and tag-based search
- **Task List Enhancements**: Added grid and list view options with improved sorting capabilities

## Features

### Backend (Go):
- RESTful API using Gin framework
- Persistent storage with BoltDB (a pure Go key-value database)
- Clean architecture with separation of concerns
- Comprehensive middleware architecture
- Task management with CRUD operations
- Task categorization, priorities, and tagging
- Task notes, reminders, and attachments
- Powerful search and filtering capabilities
- Statistics and analytics
- Batch operations support

### Frontend (React):
- Modern UI with Material-UI components
- Responsive design with mobile support
- Dashboard with task analytics and visualizations
- Task listing with advanced filtering and search
- Task creation and editing with rich options
- Form validation and error handling
- Theme customization

## Architecture

The project follows clean architecture principles with clear separation of concerns:

```
taskmanager/
├── cmd/                   # Application entry points
│   └── api/               # API server executable
│       └── main.go        # Main application entry point
├── config/                # Configuration
│   └── config.go          # Application configuration
├── internal/              # Private application code
│   ├── handlers/          # HTTP request handlers
│   │   ├── task_handler.go # Task-related handlers
│   │   └── router.go      # Route configuration
│   ├── middleware/        # HTTP middleware components
│   │   └── middleware.go  # Middleware implementations
│   ├── models/            # Data models
│   │   ├── task.go        # Task model and related types
│   │   └── user.go        # User model and authentication
│   ├── storage/           # Database interfaces and implementations
│   │   ├── storage.go     # Storage interfaces
│   │   └── boltdb.go      # BoltDB implementation
│   └── utils/             # Utility functions
│       └── utils.go       # Helper utilities
├── pkg/                   # Public libraries for external usage
├── frontend/              # React frontend application
│   ├── src/               # Frontend source code
│   │   ├── components/    # React components
│   │   ├── pages/         # Page components
│   │   ├── services/      # API communication
│   │   └── theme.ts       # UI theme configuration
│   └── public/            # Static files
└── README.md              # Project documentation
```

### Design Principles

This project adheres to the following design principles:

1. **Separation of Concerns**: Each component has a distinct responsibility
2. **Interface-Driven Design**: Components interact through well-defined interfaces
3. **Dependency Injection**: Dependencies are injected rather than instantiated directly
4. **Clean Architecture**: Inner layers don't depend on outer layers
5. **SOLID Principles**: Single responsibility, Open-closed, Liskov substitution, Interface segregation, Dependency inversion

## Tech Stack

### Backend:
- Go (Golang) 1.18+
- Gin Web Framework
- BoltDB for storage
- Middleware pattern for cross-cutting concerns

### Frontend:
- React 18
- TypeScript
- Material-UI v5
- React Router v6
- Axios for API communication
- Chart.js for data visualization

## Getting Started

### Prerequisites
- Go 1.18 or higher
- Node.js 16.x or higher
- npm 8.x or higher

### Backend Setup
1. Clone the repository
2. Navigate to the project root
3. Run the Go application:
   ```bash
   # From the project root directory
   go run cmd/api/main.go
   ```
   
   The server will start at http://localhost:8080 by default

### Frontend Setup
1. Navigate to the frontend directory:
   ```bash
   cd frontend
   ```
2. Install dependencies:
   ```bash
   npm install
   ```
3. Start the development server:
   ```bash
   npm start
   ```
4. Open http://localhost:3000 in your browser

## API Endpoints

All API endpoints are prefixed with `/api/v1`

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET    | /health | Health check endpoint |
| GET    | /api/v1/stats | Get task statistics |
| GET    | /api/v1/analytics | Get task analytics data |
| GET    | /api/v1/tasks | Get all tasks with sorting and filtering |
| GET    | /api/v1/tasks/search?q={query} | Search tasks |
| GET    | /api/v1/categories | Get all task categories |
| GET    | /api/v1/tags | Get all task tags |
| GET    | /api/v1/tasks/:id | Get a specific task |
| POST   | /api/v1/tasks | Create a new task |
| PUT    | /api/v1/tasks/:id | Update a task |
| DELETE | /api/v1/tasks/:id | Delete a task |
| POST   | /api/v1/tasks/batch/complete | Complete multiple tasks |
| POST   | /api/v1/tasks/batch/create | Create multiple tasks |
| PUT    | /api/v1/tasks/batch/update | Update multiple tasks |
| DELETE | /api/v1/tasks/batch/delete | Delete multiple tasks |
| POST   | /api/v1/tasks/:id/notes | Add a note to a task |
| PUT    | /api/v1/tasks/:id/notes/:noteId | Update a task note |
| DELETE | /api/v1/tasks/:id/notes/:noteId | Delete a task note |
| POST   | /api/v1/tasks/:id/tags | Add a tag to a task |
| DELETE | /api/v1/tasks/:id/tags/:tag | Remove a tag from a task |
| POST   | /api/v1/tasks/:id/reminders | Add a reminder to a task |
| GET    | /api/v1/tasks/:id/reminders | Get task reminders |
| DELETE | /api/v1/tasks/:id/reminders/:reminderId | Remove a reminder |
| POST   | /api/v1/tasks/:id/subtasks | Add a subtask |
| PUT    | /api/v1/tasks/:id/subtasks/:subtaskId/complete | Complete a subtask |
| PUT    | /api/v1/tasks/:id/subtasks/:subtaskId/reopen | Reopen a subtask |
| DELETE | /api/v1/tasks/:id/subtasks/:subtaskId | Delete a subtask |

## Configuration

The backend can be configured using environment variables:

- `PORT`: HTTP port (default: 8080)
- `ENV`: Environment (development/production)
- `CORS_ORIGIN`: CORS allowed origin (default: *)
- `DB_PATH`: Path to the database file (default: taskmanager.db)

## Development

### Building the backend

```bash
# From the project root
go build -o taskmanager cmd/api/main.go
```

### Building the frontend for production

```bash
cd frontend
npm run build
```

### Running tests

```bash
# Backend tests
go test ./...

# Frontend tests
cd frontend
npm test
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Acknowledgments

- Built with ❤️ using Go and React
- Material-UI for the beautiful UI components
- BoltDB for the pure Go database solution 