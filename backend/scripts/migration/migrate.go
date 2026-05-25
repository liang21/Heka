package migration

import (
	"embed"
	"fmt"
	"sort"
	"strings"

	"gorm.io/gorm"
)

//go:embed *.sql
var sqlFiles embed.FS

type migration struct {
	Version int
	Up      string
	Down    string
}

func loadMigrations() ([]migration, error) {
	var migrations []migration

	entries, err := sqlFiles.ReadDir(".")
	if err != nil {
		return nil, fmt.Errorf("failed to read sql files: %w", err)
	}

	versionMap := make(map[int]*migration)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		name := entry.Name()
		if !strings.HasSuffix(name, ".sql") {
			continue
		}

		parts := strings.Split(strings.TrimSuffix(name, ".sql"), "_")
		if len(parts) < 2 {
			continue
		}

		var version int
		_, err := fmt.Sscanf(parts[0], "%d", &version)
		if err != nil {
			continue
		}

		direction := parts[1]

		content, err := sqlFiles.ReadFile(name)
		if err != nil {
			return nil, fmt.Errorf("failed to read %s: %w", name, err)
		}

		if _, exists := versionMap[version]; !exists {
			versionMap[version] = &migration{Version: version}
		}

		switch direction {
		case "up":
			versionMap[version].Up = string(content)
		case "down":
			versionMap[version].Down = string(content)
		}
	}

	for _, m := range versionMap {
		migrations = append(migrations, *m)
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	return migrations, nil
}

func Up(db *gorm.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, m := range migrations {
			if err := tx.Exec(m.Up).Error; err != nil {
				return fmt.Errorf("failed to apply migration %d: %w", m.Version, err)
			}
		}
		return nil
	})
}

func Down(db *gorm.DB) error {
	migrations, err := loadMigrations()
	if err != nil {
		return err
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for i := len(migrations) - 1; i >= 0; i-- {
			if migrations[i].Down != "" {
				if err := tx.Exec(migrations[i].Down).Error; err != nil {
					return fmt.Errorf("failed to rollback migration %d: %w", migrations[i].Version, err)
				}
			}
		}
		return nil
	})
}
