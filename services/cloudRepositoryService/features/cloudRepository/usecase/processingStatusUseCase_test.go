package usecase

import (
	"testing"
)

// ProcessingStatusUseCase uses gorm.DB directly - tests require real DB connection.
// Minimal compilable test with t.Skip for DB-dependent behavior.
func TestProcessingStatusUseCase_RequiresRealDB(t *testing.T) {
	t.Skip("ProcessingStatusUseCase uses gorm.DB directly - tests requiring real DB should run in integration environment")
}
