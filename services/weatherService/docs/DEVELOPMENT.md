# Development Guide

Development workflow and guidelines for the Weather Data Collector Service.

**Version:** 1.0.0
**Last Updated:** 2025-11-11

## Table of Contents

- [Setup Development Environment](#setup-development-environment)
- [Running Locally](#running-locally)
- [Testing](#testing)
- [Code Organization](#code-organization)
- [Adding New Features](#adding-new-features)
- [Debugging Tips](#debugging-tips)
- [Git Workflow](#git-workflow)
- [Code Review Checklist](#code-review-checklist)

## Setup Development Environment

### Prerequisites

- Go 1.21 or higher
- MySQL 8.0+
- Redis 7.0+
- Make
- Docker (for integration tests)
- Firebase account with FCM enabled

### Installation Steps

1. **Clone Repository**
   ```bash
   git clone https://github.com/JokerTrickster/joker_backend.git
   cd joker_backend/services/weatherService
   ```

2. **Install Dependencies**
   ```bash
   go mod download
   go mod verify
   ```

3. **Setup Database**
   ```bash
   # Start MySQL
   docker run -d --name mysql \
     -e MYSQL_ROOT_PASSWORD=password \
     -e MYSQL_DATABASE=joker \
     -p 3306:3306 \
     mysql:8.0

   # Apply migrations
   mysql -h localhost -u root -ppassword joker < ../../shared/db/mysql/table.sql
   ```

4. **Setup Redis**
   ```bash
   docker run -d --name redis -p 6379:6379 redis:7.0-alpine
   ```

5. **Configure Environment**
   ```bash
   cp .env.example .env
   vim .env
   # Edit configuration
   ```

6. **Get Firebase Credentials**
   - Go to Firebase Console → Project Settings → Service Accounts
   - Generate New Private Key
   - Save as `firebase-credentials.json`
   - Update `FCM_CREDENTIALS_PATH` in `.env`

7. **Verify Setup**
   ```bash
   make test
   ```

## Running Locally

### Run Scheduler

```bash
# Build
make build-scheduler

# Run
./scheduler

# Or run directly
go run cmd/scheduler/main.go cmd/scheduler/server.go
```

### Run with Hot Reload

```bash
# Install air
go install github.com/cosmtrek/air@latest

# Run with auto-reload
air
```

### Makefile Commands

```bash
make build              # Build all binaries
make build-scheduler    # Build scheduler only
make test               # Run unit tests
make integration-test   # Run integration tests
make coverage           # Generate coverage report
make lint               # Run linter
make clean              # Clean build artifacts
```

## Testing

### Unit Tests

```bash
# Run all tests
make test

# Run specific package
go test ./features/weather/scheduler/...

# Run with verbose output
go test -v ./...

# Run specific test
go test -v -run TestWeatherSchedulerService_Start ./features/weather/scheduler/...
```

### Integration Tests

```bash
# Start dependencies
docker-compose -f docker-compose.test.yml up -d

# Run integration tests
make integration-test

# Clean up
docker-compose -f docker-compose.test.yml down
```

### Test Coverage

```bash
# Generate coverage report
make coverage

# View HTML report
go tool cover -html=coverage.out
```

### Writing Tests

**Test Structure:**
```go
func TestComponentName_MethodName(t *testing.T) {
    // Arrange
    setup()

    // Act
    result := methodUnderTest()

    // Assert
    assert.Equal(t, expected, result)
}
```

**Example:**
```go
func TestWeatherSchedulerService_ProcessAlarm(t *testing.T) {
    // Arrange
    mockRepo := &MockRepository{}
    mockCache := &MockCache{}
    scheduler := NewWeatherSchedulerService(mockRepo, ...)

    // Act
    err := scheduler.processAlarm(context.Background(), alarm)

    // Assert
    assert.NoError(t, err)
    assert.Equal(t, 1, mockRepo.UpdateLastSentCallCount)
}
```

## Code Organization

### Directory Structure

```
services/weatherService/
├── cmd/
│   ├── main.go              # API server entry point
│   └── scheduler/
│       ├── main.go          # Scheduler entry point
│       └── server.go        # Metrics server
├── features/weather/
│   ├── handler/             # HTTP handlers
│   ├── repository/          # Database access
│   ├── scheduler/           # Alarm processor
│   ├── crawler/             # Weather fetcher
│   ├── cache/               # Redis cache
│   ├── notifier/            # FCM sender
│   └── model/
│       ├── entity/          # Data models
│       └── interface/       # Interface definitions
├── pkg/
│   ├── health/              # Health checks
│   ├── metrics/             # Prometheus metrics
│   └── logger/              # Structured logging
└── tests/
    └── integration/         # Integration tests
```

### Naming Conventions

**Files:**
- `component.go` - Implementation
- `component_test.go` - Unit tests
- `example_test.go` - Example tests

**Functions:**
- Public: `PascalCase`
- Private: `camelCase`

**Interfaces:**
- Prefix with `I`: `IWeatherCache`, `ISchedulerWeatherRepository`

**Structs:**
- PascalCase: `WeatherSchedulerService`, `NaverWeatherCrawler`

## Adding New Features

### Step 1: Define Interface

```go
// features/weather/model/interface/INewFeature.go
package _interface

type INewFeature interface {
    DoSomething(ctx context.Context, param string) error
}
```

### Step 2: Implement

```go
// features/weather/newfeature/implementation.go
package newfeature

type NewFeatureImpl struct {
    dependency ISomeDependency
}

func NewNewFeature(dep ISomeDependency) *NewFeatureImpl {
    return &NewFeatureImpl{dependency: dep}
}

func (n *NewFeatureImpl) DoSomething(ctx context.Context, param string) error {
    // Implementation
    return nil
}
```

### Step 3: Add Tests

```go
// features/weather/newfeature/implementation_test.go
func TestNewFeature_DoSomething(t *testing.T) {
    // Test implementation
}
```

### Step 4: Integrate

```go
// cmd/scheduler/main.go
newFeature := newfeature.NewNewFeature(dependency)
scheduler := scheduler.NewWeatherSchedulerService(..., newFeature)
```

## Debugging Tips

### Enable Debug Logging

```bash
LOG_LEVEL=debug ./scheduler
```

### Use pprof for Profiling

```go
// Add to main.go
import _ "net/http/pprof"

go func() {
    log.Println(http.ListenAndServe("localhost:6060", nil))
}()
```

**Access profiles:**
```bash
# CPU profile
go tool pprof http://localhost:6060/debug/pprof/profile?seconds=30

# Heap profile
go tool pprof http://localhost:6060/debug/pprof/heap

# Goroutines
curl http://localhost:6060/debug/pprof/goroutine?debug=1
```

### Debug Database Queries

```go
// Enable query logging
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
    Logger: logger.Default.LogMode(logger.Info),
})
```

### Debug Redis Operations

```bash
# Monitor Redis commands
redis-cli MONITOR

# Check keys
redis-cli KEYS 'weather:*'

# Get specific key
redis-cli GET 'weather:서울'
```

## Git Workflow

### Branch Naming

- Feature: `feature/alarm-snooze`
- Bug fix: `fix/cache-leak`
- Hotfix: `hotfix/fcm-crash`

### Commit Messages

Follow Conventional Commits:

```
<type>(<scope>): <subject>

<body>

<footer>
```

**Types:**
- `feat`: New feature
- `fix`: Bug fix
- `docs`: Documentation
- `style`: Code style (formatting)
- `refactor`: Code refactoring
- `test`: Tests
- `chore`: Build/tooling

**Example:**
```
feat(scheduler): add alarm snooze feature

Allow users to snooze alarms for 10 minutes.
Adds new repository method GetSnoozeEndTime.

Closes #123
```

### Pull Request Process

1. Create feature branch
2. Make changes with tests
3. Run tests: `make test`
4. Run linter: `make lint`
5. Commit with conventional commit message
6. Push and create PR
7. Address review feedback
8. Squash and merge

## Code Review Checklist

### Functionality
- [ ] Feature works as expected
- [ ] All tests pass
- [ ] No regression in existing features

### Code Quality
- [ ] Code follows Go conventions
- [ ] No code duplication
- [ ] Proper error handling
- [ ] Interfaces used where appropriate

### Testing
- [ ] Unit tests added
- [ ] Integration tests added (if applicable)
- [ ] Code coverage > 80%
- [ ] Edge cases tested

### Documentation
- [ ] Code comments added
- [ ] README updated (if applicable)
- [ ] API documentation updated (if applicable)

### Performance
- [ ] No unnecessary allocations
- [ ] Efficient algorithms
- [ ] Database queries optimized
- [ ] No goroutine leaks

### Security
- [ ] No hardcoded credentials
- [ ] Input validation
- [ ] SQL injection prevention
- [ ] Secrets in environment variables

---

## Version History

- **v1.0.0** (2025-11-11): Initial release
