package repositorypostgres

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"pr-reviewer/internal/logging"
)

func ApplyMigrations(ctx context.Context, db *DB, migrationsDir string, logger logging.Logger) error {
	if logger == nil {
		logger = logging.StdLogger{}
	}

	entries, err := os.ReadDir(migrationsDir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}

	sort.Slice(entries, func(i, j int) bool { return entries[i].Name() < entries[j].Name() })

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}
		if filepath.Ext(entry.Name()) != ".sql" {
			continue
		}

		path := filepath.Join(migrationsDir, entry.Name())
		query, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read migration %s: %w", entry.Name(), err)
		}

		logger.Info("Applying migration %s", entry.Name())
		if _, err := db.SQL.ExecContext(ctx, string(query)); err != nil {
			logger.Error("Failed migration %s: %v", entry.Name(), err)
			return fmt.Errorf("apply migration %s: %w", entry.Name(), err)
		}
		logger.Info("Applied migration %s", entry.Name())
	}

	return nil
}
