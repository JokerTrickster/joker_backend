package repository

import (
	"testing"

	"gorm.io/gorm"
)

func requireTable(t *testing.T, db *gorm.DB, tableName string) {
	if !db.Migrator().HasTable(tableName) {
		t.Skipf("Table %s does not exist - skipping integration test", tableName)
	}
}
