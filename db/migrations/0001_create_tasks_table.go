package migrations

import "github.com/dbarzdys/jobq/db"

func init() {
	db.RegisterMigration(db.Migration{
		ID: 0001,
		Up: func() string {
			return `
				CREATE TABLE IF NOT EXISTS jobq_tasks (
					id BIGSERIAL,
					job_name varchar(100) NOT NULL,
					body jsonb NOT NULL,
					retries int NOT NULL,
					timeout timestamp,
					start_at timestamp,
					PRIMARY KEY(id)
				);
			`
		},
		Down: func() string {
			return `
				DROP TABLE IF EXISTS jobq_tasks;
			`
		},
	})
}
