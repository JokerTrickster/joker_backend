#!/bin/bash

# Database Migration Script
# Usage:
#   ./scripts/migrate.sh create <migration_name>  # Create new migration
#   ./scripts/migrate.sh up                        # Apply all pending migrations
#   ./scripts/migrate.sh down                      # Rollback last migration
#   ./scripts/migrate.sh version                   # Show current migration version
#   ./scripts/migrate.sh force <version>           # Force set migration version (use with caution)

set -e

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
fi

# Configuration
DB_HOST=${DB_HOST:-localhost}
DB_PORT=${DB_PORT:-3306}
DB_USER=${DB_USER:-joker_user}
DB_PASSWORD=${DB_PASSWORD:-joker_password}
DB_NAME=${DB_NAME:-backend_dev}
MIGRATIONS_DIR="migrations"

# Database connection string
DATABASE_URL="mysql://${DB_USER}:${DB_PASSWORD}@tcp(${DB_HOST}:${DB_PORT})/${DB_NAME}?multiStatements=true"

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Functions
print_success() {
    echo -e "${GREEN}✓ $1${NC}"
}

print_warning() {
    echo -e "${YELLOW}⚠ $1${NC}"
}

print_error() {
    echo -e "${RED}✗ $1${NC}"
}

print_info() {
    echo -e "ℹ $1"
}

show_usage() {
    cat <<EOF
Database Migration Tool

Usage:
  ./scripts/migrate.sh <command> [arguments]

Commands:
  create <name>   Create a new migration file
  up              Apply all pending migrations
  down            Rollback the last migration
  version         Show current migration version
  force <version> Force set migration version (dangerous!)
  help            Show this help message

Environment Variables:
  DB_HOST         Database host (default: localhost)
  DB_PORT         Database port (default: 3306)
  DB_USER         Database user (default: joker_user)
  DB_PASSWORD     Database password (default: joker_password)
  DB_NAME         Database name (default: backend_dev)

Examples:
  # Create a new migration
  ./scripts/migrate.sh create add_user_status

  # Apply all pending migrations
  ./scripts/migrate.sh up

  # Rollback last migration
  ./scripts/migrate.sh down

  # Check current version
  ./scripts/migrate.sh version

EOF
}

create_migration() {
    if [ -z "$1" ]; then
        print_error "Migration name is required"
        echo "Usage: ./scripts/migrate.sh create <migration_name>"
        exit 1
    fi

    MIGRATION_NAME=$1
    print_info "Creating new migration: ${MIGRATION_NAME}"

    # Create migration files using migrate CLI
    cd "${MIGRATIONS_DIR}"
    migrate create -ext sql -dir . -seq "${MIGRATION_NAME}"
    cd ..

    print_success "Migration files created in ${MIGRATIONS_DIR}/"
    echo ""
    echo "Next steps:"
    echo "1. Edit the .up.sql file to add your schema changes"
    echo "2. Edit the .down.sql file to add the rollback logic"
    echo "3. Run './scripts/migrate.sh up' to apply the migration"
}

migrate_up() {
    print_info "Applying pending migrations..."
    print_info "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

    if migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" up; then
        print_success "All migrations applied successfully"
    else
        EXIT_CODE=$?
        if [ $EXIT_CODE -eq 0 ]; then
            print_info "No new migrations to apply"
        else
            print_error "Migration failed with exit code: ${EXIT_CODE}"
            exit $EXIT_CODE
        fi
    fi
}

migrate_down() {
    print_warning "Rolling back last migration..."
    print_info "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

    if migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" down 1; then
        print_success "Migration rolled back successfully"
    else
        EXIT_CODE=$?
        print_error "Rollback failed with exit code: ${EXIT_CODE}"
        exit $EXIT_CODE
    fi
}

show_version() {
    print_info "Current migration version:"
    print_info "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

    if migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" version; then
        print_success "Version retrieved successfully"
    else
        EXIT_CODE=$?
        if [ $EXIT_CODE -eq 0 ]; then
            print_info "No migrations applied yet"
        else
            print_error "Failed to get version with exit code: ${EXIT_CODE}"
            exit $EXIT_CODE
        fi
    fi
}

force_version() {
    if [ -z "$1" ]; then
        print_error "Version number is required"
        echo "Usage: ./scripts/migrate.sh force <version>"
        exit 1
    fi

    VERSION=$1
    print_warning "⚠️  WARNING: Forcing migration version to ${VERSION}"
    print_warning "This should only be used to recover from a dirty state!"
    print_info "Database: ${DB_NAME}@${DB_HOST}:${DB_PORT}"

    read -p "Are you sure you want to continue? (yes/no): " -r
    if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
        print_info "Aborted"
        exit 0
    fi

    if migrate -path "${MIGRATIONS_DIR}" -database "${DATABASE_URL}" force "${VERSION}"; then
        print_success "Version forced to ${VERSION}"
    else
        EXIT_CODE=$?
        print_error "Force failed with exit code: ${EXIT_CODE}"
        exit $EXIT_CODE
    fi
}

# Main script
case "$1" in
    create)
        create_migration "$2"
        ;;
    up)
        migrate_up
        ;;
    down)
        migrate_down
        ;;
    version)
        show_version
        ;;
    force)
        force_version "$2"
        ;;
    help|--help|-h)
        show_usage
        ;;
    *)
        print_error "Unknown command: $1"
        echo ""
        show_usage
        exit 1
        ;;
esac
