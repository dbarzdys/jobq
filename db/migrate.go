package db

import (
	"database/sql"
	"sort"

	"github.com/jmoiron/sqlx"
)

type Migration struct {
	ID   int
	Up   func() string
	Down func() string
}

type migrationList []*Migration

func (list migrationList) Len() int {
	return len(list)
}

func (list migrationList) Less(i, j int) bool {
	return list[i].ID < list[j].ID
}

func (list migrationList) Swap(i, j int) {
	list[i], list[j] = list[j], list[i]
}

func (list migrationList) Last() *Migration {
	if list.Len() == 0 {
		return nil
	}
	return list[list.Len()-1]
}

func (list migrationList) FindByID(id int) int {
	for at, m := range list {
		if m.ID == id {
			return at
		}
	}
	return -1
}

func (list migrationList) First() *Migration {
	if list.Len() == 0 {
		return nil
	}
	return list[0]
}

var migrations migrationList

func RegisterMigration(migration Migration) {
	migrations = append(migrations, &migration)
}

type MigrationVersion struct {
	ID     int
	Active bool
}

func setupVersionTable(db *sql.DB) error {
	stmt := `
		CREATE TABLE IF NOT EXISTS jobq_version (
			id SERIAL,
			active boolean NOT NULL,
			PRIMARY KEY(id)
		);
		CREATE UNIQUE INDEX IF NOT EXISTS jobq_version_unique_active ON jobq_version (active) WHERE (active = true);
	`
	_, err := db.Exec(stmt)
	return err
}

func getActiveVersionID(db *sql.DB) (int, error) {
	stmt := `
		SELECT id
		FROM jobq_version
		WHERE active = true;
	`
	rows, err := db.Query(stmt)
	if err != nil {
		return -1, err
	}
	if !rows.Next() {
		return -1, sql.ErrNoRows
	}
	defer rows.Close()
	id := -1
	if err = rows.Scan(&id); err != nil {
		return -1, err
	}
	return id, nil
}

func removeActive(tx *sql.Tx) error {
	stmt := `
		UPDATE job_version
		SET active = false
		WHERE active = true;
	`
	_, err := tx.Exec(stmt)
	return err
}

func setActive(tx *sqlx.Tx) error {
	stmt := `
		INSERT into job_version (id, active)
		VALUES ($1, true)
		ON CONFLICT (id)
		UPDATE SET active = true;
	`
	_, err := tx.Exec(stmt)
	return err
}

func Migrate(db *sql.DB, to int) error {
	if migrations.Len() == 0 {
		return nil
	}
	sort.Sort(migrations)
	err := setupVersionTable(db)
	if err != nil {
		return err
	}
	activeID, err := getActiveVersionID(db)
	startAt := -1
	if err == sql.ErrNoRows {
		startAt = -1
	} else if err != nil {
		return err
	} else {
		startAt = migrations.FindByID(activeID)
	}
	if startAt == to {
		return nil
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if startAt > to {
		for at := startAt; at > to; at-- {
			err = down(tx, at)
			if err != nil {
				return err
			}
		}
	}
	if startAt < to {
		for at := startAt + 1; at < to; at++ {
			err = up(tx, at)
			if err != nil {
				return err
			}
		}
	}
	return err
}

func up(tx *sql.Tx, at int) error {
	migration := migrations[at]
	_, err := tx.Exec(migration.Up())
	if err != nil {
		return err
	}
	return err
}

func down(tx *sql.Tx, at int) error {
	migration := migrations[at]
	_, err := tx.Exec(migration.Down())
	if err != nil {
		return err
	}
	return err
}
