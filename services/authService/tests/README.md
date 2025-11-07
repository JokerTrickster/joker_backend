# Testing Guide for Auth Service

## Testing Philosophy

**E2E Testing Only**: This project uses End-to-End (E2E) testing exclusively. All API endpoints must have comprehensive E2E test coverage before deployment.

**Test Requirements**:
- Every new API endpoint MUST have E2E tests
- Tests must validate all success and failure scenarios
- Tests must cover edge cases and validation rules
- Tests run automatically in CI/CD before deployment

## Test Structure

### Directory Structure
```
services/auth-service/tests/
└── e2e/
    ├── setup_test.go       # Test environment setup
    ├── user_api_test.go    # User API tests
    └── ...                 # Other API test files
```

### Test Components

#### setup_test.go
- `TestMain()`: Test lifecycle management
- `setupTestEnvironment()`: Database and server initialization
- `teardownTestEnvironment()`: Cleanup after all tests
- Test fixtures: Helper functions for test data creation

#### API Test Files
- One file per API domain (e.g., `user_api_test.go`)
- Table-driven tests for multiple scenarios
- Comprehensive validation for responses

## Running Tests

### Local Development

**Run all E2E tests:**
```bash
make test-e2e
```

**Run all tests (including unit tests when added):**
```bash
make test
```

**Run with coverage report:**
```bash
make test-coverage
# Opens coverage.html in browser
```

**Watch mode (requires entr):**
```bash
make test-watch
# Auto-runs tests when Go files change
```

### CI/CD

Tests run automatically in GitHub Actions before deployment:
1. Code pushed to main/develop
2. CI detects changed service
3. **E2E tests execute** (must pass)
4. If tests pass → deployment proceeds
5. If tests fail → deployment blocked

## Writing E2E Tests

### Test Template

```go
func TestAPI_Action(t *testing.T) {
    // Clean up before test
    cleanupTestData()

    // Setup test data if needed
    testUser, err := createTestUser("Test User", "test@example.com")
    if err != nil {
        t.Fatalf("Failed to create test user: %v", err)
    }
    defer deleteTestUser(testUser.ID)

    // Define test cases
    tests := []struct {
        name           string
        requestBody    string
        expectedStatus int
        validate       func(t *testing.T, body map[string]interface{})
    }{
        {
            name:           "Success case",
            requestBody:    `{"field":"value"}`,
            expectedStatus: http.StatusOK,
            validate: func(t *testing.T, body map[string]interface{}) {
                if !body["success"].(bool) {
                    t.Error("Expected success to be true")
                }
                // More validations...
            },
        },
        {
            name:           "Failure case",
            requestBody:    `{"field":""}`,
            expectedStatus: http.StatusBadRequest,
            validate: func(t *testing.T, body map[string]interface{}) {
                if body["success"].(bool) {
                    t.Error("Expected success to be false")
                }
            },
        },
    }

    // Run test cases
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest(http.MethodPost, "/api/v1/endpoint", strings.NewReader(tt.requestBody))
            req.Header.Set("Content-Type", "application/json")
            rec := httptest.NewRecorder()

            testServer.ServeHTTP(rec, req)

            if rec.Code != tt.expectedStatus {
                t.Errorf("Expected status %d, got %d", tt.expectedStatus, rec.Code)
            }

            var responseBody map[string]interface{}
            if err := json.Unmarshal(rec.Body.Bytes(), &responseBody); err != nil {
                t.Fatalf("Failed to parse response body: %v", err)
            }

            if tt.validate != nil {
                tt.validate(t, responseBody)
            }
        })
    }
}
```

### Test Scenarios to Cover

For each API endpoint, create tests for:

**Success Cases:**
- ✅ Valid request with correct data
- ✅ Valid request with minimum required fields
- ✅ Valid request with optional fields

**Validation Errors:**
- ❌ Missing required fields
- ❌ Empty values for required fields
- ❌ Invalid data types
- ❌ Invalid data formats (e.g., invalid email)
- ❌ Invalid JSON

**Business Logic Errors:**
- ❌ Duplicate resources (e.g., duplicate email)
- ❌ Resource not found
- ❌ Invalid ID formats
- ❌ Unauthorized access (when auth is implemented)

**Edge Cases:**
- 🔍 Concurrent requests
- 🔍 Negative numbers
- 🔍 Very large numbers
- 🔍 Special characters in strings
- 🔍 Empty arrays/objects

### Example: User API Test Cases

```go
// services/auth-service/tests/e2e/user_api_test.go

// ✅ TestUserAPI_CreateUser - Success & Validation
- Success: Valid user creation
- Fail: Missing email
- Fail: Missing name
- Fail: Empty email
- Fail: Invalid JSON

// ✅ TestUserAPI_GetUser - Retrieval & Errors
- Success: Get existing user
- Fail: User not found
- Fail: Invalid user ID format
- Fail: Negative user ID

// ✅ TestUserAPI_DuplicateEmail - Business Logic
- Fail: Duplicate email constraint

// ✅ TestUserAPI_ConcurrentCreation - Concurrency
- Success: 10 concurrent user creations
```

## Test Database

### Configuration

Tests use a separate test database:
- **Production DB**: `backend_dev`
- **Test DB**: `backend_dev_test`

Environment variables for tests:
```bash
TEST_DB_HOST=localhost
TEST_DB_PORT=3306
TEST_DB_USER=joker_user
TEST_DB_PASSWORD=joker_password
TEST_DB_NAME=backend_dev_test
```

### Database Lifecycle

**Before all tests** (TestMain):
1. Connect to database
2. Create test database if not exists
3. Run migrations to create tables
4. Setup test server

**Before each test**:
1. Call `cleanupTestData()` to clear all test data
2. Create specific test fixtures as needed

**After each test**:
1. Clean up created fixtures (using defer)

**After all tests** (TestMain):
1. Clean all test data
2. Close database connection

### Test Fixtures

Use provided helper functions:

```go
// Create test user
testUser, err := createTestUser("John Doe", "john@example.com")
if err != nil {
    t.Fatalf("Failed to create test user: %v", err)
}
defer deleteTestUser(testUser.ID) // Clean up after test

// Count users (for validation)
count, err := countUsers()
if err != nil {
    t.Fatalf("Failed to count users: %v", err)
}

// Clean all test data
cleanupTestData()
```

## Best Practices

### Test Organization
- ✅ One test file per API domain
- ✅ Group related tests using table-driven approach
- ✅ Use descriptive test names: `TestAPI_Action_Scenario`
- ✅ Clean up test data before each test

### Test Data Management
- ✅ Use test fixtures for common data creation
- ✅ Clean up created data using `defer`
- ✅ Avoid hardcoded IDs, use fixtures instead
- ✅ Isolate tests - don't depend on other test data

### Assertions
- ✅ Check HTTP status codes
- ✅ Validate response structure
- ✅ Verify all important fields in response
- ✅ Test error messages are meaningful
- ✅ Use `t.Errorf` for failures, `t.Fatalf` for fatal errors

### Test Independence
- ✅ Each test should run independently
- ✅ Tests should not depend on execution order
- ✅ Clean state before each test
- ✅ No shared mutable state between tests

### Performance
- ✅ Tests should run quickly (<5 seconds per test)
- ✅ Use `t.Parallel()` for independent tests (when appropriate)
- ✅ Clean up efficiently - batch operations when possible
- ✅ Avoid unnecessary database calls

## Common Patterns

### Testing Create Operations
```go
// POST request
req := httptest.NewRequest(http.MethodPost, "/api/v1/users", strings.NewReader(requestBody))
req.Header.Set("Content-Type", "application/json")
rec := httptest.NewRecorder()
testServer.ServeHTTP(rec, req)

// Validate response
if rec.Code != http.StatusCreated {
    t.Errorf("Expected status %d, got %d", http.StatusCreated, rec.Code)
}

// Parse and validate body
var responseBody map[string]interface{}
json.Unmarshal(rec.Body.Bytes(), &responseBody)

data := responseBody["data"].(map[string]interface{})
if data["id"] == nil {
    t.Error("Expected ID to be set")
}
```

### Testing Get Operations
```go
// Create test data first
testUser, _ := createTestUser("Jane Doe", "jane@example.com")
defer deleteTestUser(testUser.ID)

// GET request
req := httptest.NewRequest(http.MethodGet, fmt.Sprintf("/api/v1/users/%d", testUser.ID), nil)
rec := httptest.NewRecorder()
testServer.ServeHTTP(rec, req)

// Validate response
if rec.Code != http.StatusOK {
    t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
}
```

### Testing Concurrent Operations
```go
numRequests := 10
done := make(chan bool, numRequests)

for i := 0; i < numRequests; i++ {
    go func(index int) {
        // Perform request
        // ...
        done <- true
    }(i)
}

// Wait for all requests
for i := 0; i < numRequests; i++ {
    <-done
}

// Validate results
count, _ := countUsers()
if count != numRequests {
    t.Errorf("Expected %d users, got %d", numRequests, count)
}
```

## Troubleshooting

### Database Connection Issues
```bash
# Check MySQL is running
docker ps | grep mysql

# Check database exists
mysql -u joker_user -p -e "SHOW DATABASES;"

# Verify test database
mysql -u joker_user -p backend_dev_test -e "SHOW TABLES;"
```

### Test Failures
1. **Check error messages**: Read the full error output
2. **Verify test data**: Ensure test database is clean
3. **Check migrations**: Verify tables are created correctly
4. **Review logs**: Set `LOG_LEVEL=debug` for verbose output

### Coverage Issues
```bash
# Generate coverage report
make test-coverage

# View coverage in browser
open coverage.html

# Target: >80% coverage for all handlers
```

## CI/CD Integration

### GitHub Actions Workflow

Tests are executed automatically:

```yaml
- name: Run E2E Tests
  env:
    TEST_DB_HOST: localhost
    TEST_DB_PORT: 3306
    TEST_DB_USER: ${{ secrets.DB_USER }}
    TEST_DB_PASSWORD: ${{ secrets.DB_PASSWORD }}
    TEST_DB_NAME: backend_dev_test
  run: |
    cd services/auth-service
    go mod download
    make test-e2e
```

### Test Failures in CI

When tests fail in CI:
1. ❌ Deployment is blocked
2. 🔍 Check GitHub Actions logs
3. 🐛 Fix failing tests locally
4. ✅ Push fix
5. 🚀 CI re-runs automatically

## Adding Tests for New Features

**Checklist for new API endpoints:**

1. ✅ Create test function in appropriate file
2. ✅ Cover all success scenarios
3. ✅ Cover all validation errors
4. ✅ Cover business logic errors
5. ✅ Cover edge cases
6. ✅ Run tests locally: `make test-e2e`
7. ✅ Verify coverage: `make test-coverage`
8. ✅ Commit tests WITH implementation code
9. ✅ CI runs tests automatically
10. ✅ Deployment proceeds if tests pass

## Resources

- **httptest Package**: https://pkg.go.dev/net/http/httptest
- **Echo Testing**: https://echo.labstack.com/docs/testing
- **Table-Driven Tests**: https://go.dev/wiki/TableDrivenTests
- **Go Testing**: https://go.dev/doc/tutorial/add-a-test

## Questions?

For testing questions or issues:
1. Check this documentation
2. Review existing test examples in `tests/e2e/`
3. Consult team members
4. Update this documentation when solutions are found
