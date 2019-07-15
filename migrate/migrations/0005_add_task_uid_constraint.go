package migrations

import (
	"github.com/dbarzdys/jobq/migrate"
)

func init() {
	migrate.RegisterMigration(migrate.Migration{
		ID: 0005,
		Up: func() string {
			return `
				ALTER TABLE jobq_tasks
				ADD CONSTRAINT jobq_task_uid_unique UNIQUE (uid);
			`
		},
		Down: func() string {
			return `
				ALTER TABLE jobq_tasks
				DROP CONSTRAINT jobq_task_uid_unique;
			`
		},
	})
}
