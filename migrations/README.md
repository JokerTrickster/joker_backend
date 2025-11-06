# Database Migrations

This directory contains database migration files for the Joker Backend project. Migrations are managed using [golang-migrate](https://github.com/golang-migrate/migrate).

## Overview

**Migration System**: golang-migrate v4
**Database**: MySQL 8.0
**Auto-Migration**: Enabled on service startup
**Rollback Support**: Yes (down migrations)

## Migration Files

Migrations follow a sequential naming convention:

```
000001_create_users_table.up.sql      # Apply migration
000001_create_users_table.down.sql    # Rollback migration
000002_add_user_status.up.sql         # Next migration
000002_add_user_status.down.sql       # Rollback
...
```

**File Naming Pattern**: `{version}_{description}.{up|down}.sql`

- `version`: Sequential number with leading zeros (000001, 000002, etc.)
- `description`: Snake_case description of the migration
- `up`: SQL to apply the migration
- `down`: SQL to rollback the migration

## Usage

### CLI Tool

Use the migration script for manual migration management:

```bash
# Create a new migration
./scripts/migrate.sh create add_user_status

# Apply all pending migrations
./scripts/migrate.sh up

# Rollback last migration
./scripts/migrate.sh down

# Check current version
./scripts/migrate.sh version

# Force version (recovery only)
./scripts/migrate.sh force 1
```

### Environment Variables

Configure database connection:

```bash
DB_HOST=localhost        # Database host
DB_PORT=3306            # Database port
DB_USER=joker_user      # Database user
DB_PASSWORD=***         # Database password
DB_NAME=backend_dev     # Database name
```

### Auto-Migration

Migrations run automatically when services start:

**Local Development**:
```bash
cd services/auth-service
go run cmd/server/main.go
# Migrations applied automatically
```

**Docker**:
```bash
docker-compose up
# Migrations applied on container start
```

**Production Deployment**:
```bash
# Migrations applied automatically during deployment
./scripts/deploy-service.sh auth-service 6000
```

## Creating Migrations

### Step 1: Create Migration Files

```bash
./scripts/migrate.sh create add_user_profile
```

This creates:
- `migrations/000002_add_user_profile.up.sql`
- `migrations/000002_add_user_profile.down.sql`

### Step 2: Write Up Migration

Edit `000002_add_user_profile.up.sql`:

```sql
-- Add profile fields to users table
ALTER TABLE users
ADD COLUMN phone VARCHAR(20),
ADD COLUMN bio TEXT,
ADD COLUMN avatar_url VARCHAR(500);

-- Add index for phone lookup
CREATE INDEX idx_phone ON users(phone);
```

### Step 3: Write Down Migration

Edit `000002_add_user_profile.down.sql`:

```sql
-- Remove profile fields from users table
ALTER TABLE users
DROP COLUMN phone,
DROP COLUMN bio,
DROP COLUMN avatar_url;

-- Drop phone index
DROP INDEX idx_phone ON users;
```

### Step 4: Test Migration

```bash
# Apply migration
./scripts/migrate.sh up

# Verify database schema
mysql -u joker_user -p backend_dev -e "DESCRIBE users;"

# Test rollback
./scripts/migrate.sh down

# Verify rollback worked
mysql -u joker_user -p backend_dev -e "DESCRIBE users;"

# Re-apply for production
./scripts/migrate.sh up
```

### Step 5: Commit

```bash
git add migrations/
git commit -m "Add user profile fields migration"
git push
```

## Best Practices

### ✅ DO

**Write Reversible Migrations**
- Every `up` migration MUST have a corresponding `down` migration
- Test both up and down migrations before committing

**Keep Migrations Small**
- One logical change per migration
- Easier to review, test, and rollback

**Use Descriptive Names**
- `add_user_status` ✅
- `update_users` ❌ (too vague)

**Add Comments**
- Explain why the migration is needed
- Document any complex SQL logic

**Test Thoroughly**
```bash
# Test sequence
1. Apply migration: ./scripts/migrate.sh up
2. Verify schema and data
3. Rollback: ./scripts/migrate.sh down
4. Verify rollback worked
5. Re-apply: ./scripts/migrate.sh up
```

**Handle Data Migration Safely**
```sql
-- Good: Update with default for existing rows
ALTER TABLE users
ADD COLUMN status VARCHAR(20) DEFAULT 'active' NOT NULL;

-- Update existing data if needed
UPDATE users SET status = 'active' WHERE status IS NULL;
```

### ❌ DON'T

**Don't Modify Existing Migrations**
- Once a migration is committed and deployed, DON'T edit it
- Create a new migration to make changes

**Don't Use DROP without IF EXISTS**
```sql
-- Bad
DROP TABLE users;

-- Good
DROP TABLE IF EXISTS users;
```

**Don't Skip Down Migrations**
```sql
-- Bad
-- down.sql is empty or just has a comment

-- Good
DROP TABLE IF EXISTS new_table;
ALTER TABLE old_table DROP COLUMN new_column;
```

**Don't Commit Broken Migrations**
- Always test migrations locally before committing
- Verify both up and down work correctly

**Don't Mix Schema and Data Changes**
```sql
-- Bad: Schema + Data in one migration
ALTER TABLE users ADD COLUMN role VARCHAR(20);
UPDATE users SET role = 'user';

-- Good: Separate migrations for clarity
-- Migration 1: Add column with default
-- Migration 2: Update data if complex logic needed
```

## Migration Workflow

### Development Flow

```
1. Create migration: ./scripts/migrate.sh create <name>
2. Write SQL in .up.sql and .down.sql files
3. Test locally: ./scripts/migrate.sh up
4. Test rollback: ./scripts/migrate.sh down
5. Re-apply: ./scripts/migrate.sh up
6. Commit migration files
7. Push to repository
8. CI/CD runs migrations automatically
```

### CI/CD Flow

```
1. Code pushed to main/develop
2. CI detects changed service
3. E2E tests run (with migrations)
4. Tests pass
5. Deployment starts
6. Service starts → migrations auto-run
7. Service becomes available
```

## Migration States

### Clean State
```
Current version: 3
Dirty: false
```
Everything is normal. Safe to apply new migrations.

### Dirty State
```
Current version: 3
Dirty: true
```

Migration failed mid-execution. Database might be in inconsistent state.

**Recovery**:
```bash
# 1. Check what went wrong
./scripts/migrate.sh version

# 2. Manually fix database if needed
mysql -u joker_user -p backend_dev

# 3. Force version to clean state
./scripts/migrate.sh force 3

# 4. Try again
./scripts/migrate.sh up
```

## Testing Migrations

### Unit Testing

Migrations are tested automatically in E2E tests:

```go
// tests/e2e/setup_test.go
func setupTestEnvironment() error {
    // Migrations applied automatically
    migrateConfig := migrate.Config{
        MigrationsPath: "../../../migrations",
        DatabaseName:   "backend_dev_test",
    }
    migrate.Run(testDB.DB, migrateConfig)
}
```

### Manual Testing

```bash
# 1. Create test database
mysql -u joker_user -p -e "CREATE DATABASE test_migrations;"

# 2. Test migrations
DB_NAME=test_migrations ./scripts/migrate.sh up

# 3. Verify schema
mysql -u joker_user -p test_migrations -e "SHOW TABLES;"

# 4. Test rollback
DB_NAME=test_migrations ./scripts/migrate.sh down

# 5. Clean up
mysql -u joker_user -p -e "DROP DATABASE test_migrations;"
```

## Common Issues

### Issue: "dirty database version"

**Cause**: Migration failed mid-execution

**Solution**:
```bash
# Check current state
./scripts/migrate.sh version

# Manually inspect and fix database
mysql -u joker_user -p backend_dev

# Force to last known good version
./scripts/migrate.sh force <version>
```

### Issue: "no change"

**Cause**: All migrations already applied

**Solution**: This is normal. No action needed.

### Issue: "file does not exist"

**Cause**: Migration files not found

**Solution**:
```bash
# Check migrations directory exists
ls -la migrations/

# Verify MIGRATIONS_PATH environment variable
echo $MIGRATIONS_PATH

# Update path if needed
export MIGRATIONS_PATH=./migrations
```

### Issue: "connection refused"

**Cause**: Database not running or wrong connection info

**Solution**:
```bash
# Check MySQL is running
docker ps | grep mysql
# OR
lsof -i :3306

# Verify connection settings
echo $DB_HOST $DB_PORT $DB_USER
```

## Migration History

Track applied migrations:

```bash
# Check current version
./scripts/migrate.sh version

# View migration history in database
mysql -u joker_user -p backend_dev -e "SELECT * FROM schema_migrations;"
```

## Rollback Strategy

### Immediate Rollback (Same Day)
```bash
# Rollback last migration
./scripts/migrate.sh down

# Redeploy previous version
git revert <commit>
git push
```

### Planned Rollback (Data Preservation)
```sql
-- Instead of DROP, use RENAME
-- migration_up.sql
ALTER TABLE users ADD COLUMN new_field VARCHAR(100);

-- migration_down.sql (preserves data)
ALTER TABLE users DROP COLUMN new_field;
-- OR for data preservation
ALTER TABLE users_backup SELECT * FROM users;
ALTER TABLE users DROP COLUMN new_field;
```

## Directory Structure

```
migrations/
├── README.md                           # This file
├── 000001_create_users_table.up.sql   # Users table creation
├── 000001_create_users_table.down.sql # Users table rollback
├── 000002_*.up.sql                    # Future migrations
├── 000002_*.down.sql
└── ...
```

## Additional Resources

- [golang-migrate documentation](https://github.com/golang-migrate/migrate)
- [MySQL ALTER TABLE](https://dev.mysql.com/doc/refman/8.0/en/alter-table.html)
- [Migration best practices](https://www.brunton.dev/posts/2020-03-16-database-migrations-best-practices/)

## Support

For migration issues:
1. Check this documentation
2. Review error messages carefully
3. Test in local environment first
4. Consult team before forcing versions in production
