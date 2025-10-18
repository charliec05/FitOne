# FitONEX Backend

A production-grade Go backend for the FitONEX fitness super app.

## Features

- **Authentication**: JWT-based user authentication
- **User Management**: Registration, login, profile management
- **Workout CRUD**: Create, read, update, delete workouts
- **Cursor-based Pagination**: Efficient workout listing
- **PostgreSQL**: Robust database with proper indexing
- **Redis**: Caching layer for improved performance
- **Docker**: Containerized deployment
- **Clean Architecture**: Modular, testable code structure

## Tech Stack

- **Language**: Go 1.21
- **Router**: Chi (lightweight, idiomatic HTTP router)
- **Database**: PostgreSQL 15
- **Cache**: Redis 7
- **Authentication**: JWT tokens
- **Containerization**: Docker & Docker Compose

## Project Structure

```
backend/
├── cmd/api/                 # Application entrypoint
├── internal/                # Private application code
│   ├── auth/               # JWT authentication
│   ├── config/             # Configuration management
│   ├── handlers/           # HTTP request handlers
│   ├── models/             # Data models
│   ├── server/             # HTTP server setup
│   └── store/              # Database layer
│       ├── migrations/     # Database migrations
│       ├── users/          # User data access
│       └── workouts/       # Workout data access
├── pkg/                    # Public packages
│   └── utils/              # Utility functions
├── migrations/             # SQL migration files
├── infra/                  # Infrastructure as code
├── Dockerfile             # Container definition
├── Makefile               # Build automation
└── go.mod                 # Go module definition
```

## Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Make (optional, for convenience commands)

### Development Setup

1. **Clone and setup**:
   ```bash
   cd backend
   cp env.example .env
   ```

2. **Start services**:
   ```bash
   make setup
   ```

3. **Run the server**:
   ```bash
   make run
   ```

The API will be available at `http://localhost:8080`

### Manual Setup

1. **Start database services**:
   ```bash
   cd infra
   docker-compose up -d postgres redis
   ```

2. **Run migrations**:
   ```bash
   make migrate-up
   ```

3. **Start the server**:
   ```bash
   go run cmd/api/main.go
   ```

### Seed Development Data

After the database is up and migrations are applied you can preload reference data (Seattle gyms, machines, prices, and sample reviews):

```bash
make seed-dev
```

This command is idempotent and safe to run multiple times.

## API Endpoints

### Health Check
- `GET /healthz` - Returns server status

### Authentication
- `POST /v1/auth/register` - User registration
- `POST /v1/auth/login` - User login

### User Management
- `GET /v1/profile` - Get user profile (requires auth)
- `PUT /v1/profile` - Update user profile (requires auth)

### Workouts
- `GET /v1/workouts` - List user workouts (cursor-based pagination)
- `POST /v1/workouts` - Create new workout (requires auth)
- `GET /v1/workouts/{id}` - Get specific workout (requires auth)
- `PUT /v1/workouts/{id}` - Update workout (requires auth)
- `DELETE /v1/workouts/{id}` - Delete workout (requires auth)

### Gyms & Map Data
- `GET /v1/gyms/nearby` - Nearby gyms ordered by distance (cursor pagination)

### Instruction Videos
- `POST /v1/videos/upload-url` - Request presigned URLs for uploading video + thumbnail (requires auth)
- `POST /v1/videos/finalize` - Persist video metadata after upload (requires auth)
- `GET /v1/videos` - List machine instruction videos (cursor pagination)

### Check-ins & Streaks
- `POST /v1/checkins/today` - Idempotent daily check-in (requires auth)
- `GET /v1/checkins/me` - Current and longest streak stats (requires auth)

### Exercises
- `POST /v1/exercises` - Log an exercise with sets for a given day (requires auth)
- `GET /v1/exercises` - Day view with cursor pagination (requires auth)

## Sample cURL

```bash
# Nearby gyms
curl "http://localhost:8080/v1/gyms/nearby?lat=47.61&lng=-122.33&radius_km=5&limit=10"

# Request upload URL for a video (replace $TOKEN and payload values)
curl -X POST http://localhost:8080/v1/videos/upload-url \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "machine_id": "55555555-0000-0000-0000-000000000001",
        "title": "Leg Press Basics",
        "description": "How to set up the machine",
        "content_type": "video/mp4",
        "bytes": 52428800
      }'

# Finalize an uploaded video
curl -X POST http://localhost:8080/v1/videos/finalize \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
        "machine_id": "55555555-0000-0000-0000-000000000001",
        "title": "Leg Press Basics",
        "video_key": "videos/<key-from-presign>",
        "thumb_key": "videos/thumbs/<thumb-key>",
        "duration_sec": 95
      }'

# Record today's check-in
curl -X POST http://localhost:8080/v1/checkins/today \
  -H "Authorization: Bearer $TOKEN"

# Fetch streak stats
curl http://localhost:8080/v1/checkins/me \
  -H "Authorization: Bearer $TOKEN"

# Fetch exercises for a day
curl "http://localhost:8080/v1/exercises?day=2024-05-20&limit=10" \
  -H "Authorization: Bearer $TOKEN"
```

## Cursor-based Pagination

The `/v1/workouts` endpoint supports cursor-based pagination:

```
GET /v1/workouts?cursor=2024-01-01T12:00:00Z&limit=20
```

- `cursor`: ISO timestamp for pagination (optional)
- `limit`: Number of items per page (default: 20, max: 100)

Response includes a `next` field with the cursor for the next page.

## Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `PORT` | Server port | `8080` |
| `DATABASE_URL` | PostgreSQL connection string | `postgres://fitonex:password@localhost:5432/fitonex?sslmode=disable` |
| `JWT_SECRET` | JWT signing secret | `your-super-secret-jwt-key-change-in-production` |
| `REDIS_URL` | Redis connection string | `redis://localhost:6379` |
| `MAPS_API_KEY` | Google Maps API key (optional for map tiles) | *(empty)* |
| `S3_ENDPOINT` | MinIO/S3 endpoint for media uploads | `http://localhost:9000` |
| `S3_BUCKET` | Bucket name for storing videos | `fitonex` |
| `S3_REGION` | S3 region | `us-east-1` |
| `MAX_UPLOAD_MB` | Max video upload size in MB | `100` |
| `ENVIRONMENT` | Environment (development/production) | `development` |

## Database Schema

### Users Table
```sql
CREATE TABLE users (
    id UUID PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

### Workouts Table
```sql
CREATE TABLE workouts (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    duration INTEGER NOT NULL,
    type VARCHAR(100) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
```

## Available Commands

```bash
make help          # Show available commands
make run           # Run development server
make build         # Build the application
make test          # Run tests
make lint          # Run go vet on the codebase
make seed-dev      # Seed the development database with fixtures
make clean         # Clean build artifacts
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make setup         # Complete development setup
make stop          # Stop all services
make logs          # View application logs
```

## Development

### Adding New Features

1. **Models**: Add to `internal/models/`
2. **Database**: Add to `internal/store/`
3. **Handlers**: Add to `internal/handlers/`
4. **Routes**: Update `internal/server/server.go`

### Database Migrations

1. Create SQL files in `migrations/`
2. Run `make migrate-up` to apply
3. Run `make migrate-down` to rollback

### Testing

```bash
make test
```

## Production Deployment

1. **Build Docker image**:
   ```bash
   make docker-build
   ```

2. **Deploy with Docker Compose**:
   ```bash
   cd infra
   docker-compose up -d
   ```

3. **Set production environment variables**:
   - Update `JWT_SECRET` with a secure random string
   - Configure `DATABASE_URL` for production database
   - Set `ENVIRONMENT=production`

## Security Considerations

- JWT tokens expire after 24 hours
- Passwords are hashed using bcrypt
- CORS is configured (adjust for production)
- Database connections use SSL in production
- Input validation on all endpoints

## Performance Optimizations

- Database indexes on frequently queried columns
- Cursor-based pagination for efficient large dataset handling
- Connection pooling for database operations
- Redis caching layer (ready for implementation)

## Next Steps

- [ ] Add comprehensive test suite
- [ ] Implement Redis caching
- [ ] Add request rate limiting
- [ ] Implement API documentation (OpenAPI/Swagger)
- [ ] Add monitoring and logging
- [ ] Set up CI/CD pipeline
