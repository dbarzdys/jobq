package migrations

import (
	"github.com/dbarzdys/jobq/migrate"
)

func init() {
	migrate.RegisterMigration(migrate.Migration{
		ID: 0004,
		Up: func() string {
			return `
				ALTER TABLE jobq_tasks
				ADD COLUMN uid uuid NOT NULL;
			`
		},
		Down: func() string {
			return `
				ALTER TABLE jobq_tasks
				DROP COLUMN IF EXISTS uid;
			`
		},
	})
}
