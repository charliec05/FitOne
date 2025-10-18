# FitONEX - Fitness Super App

A production-grade fitness super app with Go backend and Kotlin Android app that supports:

- **Nearby Gyms Map** with prices, machines, ratings & comments
- **Instruction Videos** for machine how-to (creator + user uploads)
- **Streak Check-ins** like Duolingo, with daily exercise logs
- **User Authentication** with JWT
- **Workout CRUD** with cursor-based pagination
- **PostgreSQL** database with Redis caching

## 🏗️ Architecture

```
FitONEX/
├── backend/                 # Go backend API
│   ├── cmd/api/            # Application entrypoint
│   ├── internal/           # Private application code
│   │   ├── auth/          # JWT authentication
│   │   ├── config/        # Configuration management
│   │   ├── handlers/      # HTTP request handlers
│   │   ├── models/        # Data models
│   │   ├── server/        # HTTP server setup
│   │   ├── store/         # Database layer
│   │   └── storage/       # S3/MinIO storage
│   ├── migrations/         # SQL migration files
│   ├── infra/             # Infrastructure as code
│   └── Dockerfile         # Container definition
├── mobile-android/         # Kotlin Android app
│   ├── app/               # Android application
│   │   ├── src/main/java/com/fitonex/app/
│   │   │   ├── data/      # Data layer (models, repositories)
│   │   │   ├── network/   # API service and interceptors
│   │   │   └── ui/        # UI layer (screens, navigation)
│   │   └── build.gradle.kts
│   └── build.gradle.kts
└── infra/                  # Infrastructure services
    └── docker-compose.yml # PostgreSQL, Redis, MinIO
```

## 🚀 Quick Start

### Prerequisites

- Go 1.21+
- Docker & Docker Compose
- Android Studio (for mobile development)
- Make (optional, for convenience commands)

### Backend Setup

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

### Android Setup

1. **Open in Android Studio**:
   ```bash
   cd mobile-android
   # Open in Android Studio
   ```

2. **Configure Google Maps**:
   - Get a Google Maps API key
   - Update `app/src/main/res/values/strings.xml`
   - Replace `YOUR_GOOGLE_MAPS_API_KEY` with your actual key

3. **Build and run**:
   - Connect Android device or start emulator
   - Click "Run" in Android Studio

## 📱 Features

### 1. Nearby Gyms Map
- **Google Maps integration** with gym markers
- **Geosearch** using Haversine formula (5km default radius)
- **Gym details**: prices, machines, ratings, reviews
- **Filters**: radius, body part, price range
- **Bottom sheet** with gym information

### 2. Instruction Videos
- **Machine-specific videos** with search and filtering
- **Upload flow**: presigned S3 URLs for video uploads
- **Video metadata**: duration, thumbnails, likes
- **User-generated content** with creator attribution

### 3. Streak Check-ins
- **Daily check-ins** with streak tracking
- **Exercise logging** with sets, reps, weight, RPE
- **Gym/machine context** for exercises
- **Progress tracking** with current and longest streaks

## 🔧 API Endpoints

### Authentication
- `POST /v1/auth/register` - User registration
- `POST /v1/auth/login` - User login

### Gyms
- `GET /v1/gyms/nearby?lat=&lng=&radius_km=&limit=` - Nearby gyms
- `GET /v1/gyms/{id}` - Gym details
- `GET /v1/gyms/{id}/machines` - Gym machines
- `GET /v1/gyms/{id}/prices` - Gym pricing
- `GET /v1/gyms/{id}/reviews` - Gym reviews (paginated)
- `POST /v1/gyms/{id}/reviews` - Create review (auth required)

### Machines
- `GET /v1/machines?query=&body_part=&limit=` - Search machines
- `GET /v1/machines/body-parts` - Get body parts
- `GET /v1/machines/{id}` - Machine details

### Videos
- `GET /v1/videos?machine_id=&limit=&cursor=` - Machine videos (paginated)
- `GET /v1/videos/{id}` - Video details
- `POST /v1/videos/upload-url` - Get presigned upload URL (auth required)
- `POST /v1/videos/finalize` - Finalize video upload (auth required)
- `POST /v1/videos/{id}/like` - Like video (auth required)
- `DELETE /v1/videos/{id}/like` - Unlike video (auth required)

### Check-ins
- `POST /v1/checkins/today` - Check in today (auth required)
- `GET /v1/checkins/me` - Get streak stats (auth required)

### Exercises
- `POST /v1/exercises` - Create exercise (auth required)
- `GET /v1/exercises?day=YYYY-MM-DD&limit=&cursor=` - Get exercises (auth required)
- `GET /v1/exercises/{id}` - Exercise details (auth required)

## 🗄️ Database Schema

### Core Tables
- **users**: User accounts with authentication
- **gyms**: Gym locations with coordinates
- **machines**: Available gym equipment
- **gym_machines**: Junction table (gym ↔ machine)
- **gym_prices**: Membership pricing plans
- **gym_reviews**: User reviews and ratings

### Video System
- **instruction_videos**: Machine instruction videos
- **video_likes**: User likes for videos

### Check-in System
- **checkins**: Daily check-ins for streak tracking
- **exercises**: Exercise sessions with context
- **sets**: Individual sets within exercises

## 🐳 Infrastructure

### Docker Services
- **PostgreSQL 15**: Primary database
- **Redis 7**: Caching layer
- **MinIO**: S3-compatible storage for videos
- **API Server**: Go backend

### Environment Variables
```bash
# Server
PORT=8080
ENVIRONMENT=development

# Database
DATABASE_URL=postgres://fitonex:password@localhost:5432/fitonex?sslmode=disable

# JWT
JWT_SECRET=your-super-secret-jwt-key-change-in-production

# Redis
REDIS_URL=redis://localhost:6379

# Features
MAPS_API_KEY=your-google-maps-api-key
S3_ENDPOINT=http://localhost:9000
S3_BUCKET=fitonex
S3_REGION=us-east-1
MAX_UPLOAD_MB=100
```

## 📱 Android App Structure

### Screens
- **Home**: Dashboard with streak stats and quick actions
- **Map**: Google Maps with nearby gyms and filters
- **Log**: Check-in and exercise logging
- **Learn**: Machine videos and upload functionality
- **Profile**: User profile and settings

### Navigation
- **Bottom Navigation** with 5 tabs
- **Material 3** design system
- **Jetpack Compose** UI framework

### Data Layer
- **Repository pattern** for data access
- **Retrofit** for API communication
- **Kotlinx Serialization** for JSON handling
- **SharedPreferences** for auth token storage

## 🛠️ Development Commands

### Backend
```bash
make help          # Show available commands
make run           # Run development server
make build         # Build the application
make test          # Run tests
make clean         # Clean build artifacts
make migrate-up    # Run database migrations
make migrate-down  # Rollback migrations
make seed          # Seed sample data
make docker-build  # Build Docker image
make docker-run    # Run with Docker Compose
make setup         # Complete development setup
make stop          # Stop all services
make logs          # View application logs
```

### Android
```bash
# In Android Studio
./gradlew assembleDebug     # Build debug APK
./gradlew test             # Run unit tests
./gradlew connectedAndroidTest  # Run instrumented tests
```

## 🔒 Security Features

- **JWT Authentication** with 24-hour expiration
- **Password hashing** using bcrypt
- **CORS configuration** for cross-origin requests
- **Input validation** on all endpoints
- **Rate limiting** for uploads and reviews
- **Content type validation** for file uploads

## 🚀 Production Deployment

### Backend
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
   - Update `JWT_SECRET` with secure random string
   - Configure `DATABASE_URL` for production database
   - Set `ENVIRONMENT=production`
   - Configure `S3_ENDPOINT` for production S3

### Android
1. **Generate signed APK**:
   - Create keystore for signing
   - Configure build variants
   - Generate release APK

2. **Publish to Google Play**:
   - Follow Google Play Console guidelines
   - Configure app signing
   - Upload and publish

## 📊 Performance Optimizations

- **Database indexes** on frequently queried columns
- **Cursor-based pagination** for efficient large dataset handling
- **Connection pooling** for database operations
- **Redis caching** layer (ready for implementation)
- **S3 presigned URLs** for efficient file uploads
- **Haversine formula** for geospatial queries

## 🧪 Testing

### Backend Testing
```bash
make test                    # Run all tests
go test ./internal/handlers  # Test handlers
go test ./internal/store     # Test data layer
```

### Android Testing
```bash
./gradlew test                    # Unit tests
./gradlew connectedAndroidTest   # Integration tests
```

## 📈 Monitoring & Logging

- **Health check endpoint**: `GET /healthz`
- **Structured logging** with request/response tracking
- **Error handling** with proper HTTP status codes
- **Database query logging** for performance monitoring

## 🔄 Cursor Pagination

All list endpoints support cursor-based pagination:

```
GET /v1/gyms/nearby?lat=40.7128&lng=-74.0060&radius_km=5&limit=20
GET /v1/videos?machine_id=123&limit=20&cursor=2024-01-01T12:00:00Z
```

- `cursor`: Base64-encoded timestamp for pagination
- `limit`: Number of items per page (max 50)
- Response includes `next` field with cursor for next page

## 🎯 Next Steps

- [ ] Add comprehensive test suite
- [ ] Implement Redis caching
- [ ] Add request rate limiting
- [ ] Implement API documentation (OpenAPI/Swagger)
- [ ] Add monitoring and logging
- [ ] Set up CI/CD pipeline
- [ ] Add push notifications
- [ ] Implement social features (follow users, share workouts)
- [ ] Add premium features and subscription management

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🤝 Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new functionality
5. Submit a pull request

## 📞 Support

For support and questions:
- Create an issue in the repository
- Check the documentation
- Review the API endpoints

---

**FitONEX** - Your complete fitness companion! 🔥💪