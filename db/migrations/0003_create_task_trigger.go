package migrations

import "github.com/dbarzdys/jobq/db"

func init() {
	db.RegisterMigration(db.Migration{
		ID: 0003,
		Up: func() string {
			return `
				DO $$ BEGIN
					IF NOT EXISTS(SELECT *
						FROM information_schema.triggers
						WHERE event_object_table = 'jobq_tasks'
						AND trigger_name = 'jobq_task_trigger'
						)
						THEN
							CREATE TRIGGER jobq_task_trigger
								AFTER INSERT ON jobq_tasks
								FOR EACH ROW EXECUTE PROCEDURE jobq_notify_task_created();
						
						END IF ;
					END;
				$$
			`
		},
		Down: func() string {
			return `
				DROP TRIGGER IF EXISTS jobq_task_trigger ON jobq_tasks;
			`
		},
	})
}
