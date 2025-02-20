//go:build !inmemory
// +build !inmemory

package repository

import (
	"database/sql"
	"errors"
	"sync"
	"time"

	"github.com/ZeroBl21/cli/ch10/pomo/pomodoro"
	_ "github.com/mattn/go-sqlite3"
)

const createTableInterval string = `
CREATE TABLE IF NOT EXISTS "interval" (
	"id" INTEGER,
	"start_time" DATETIME NOT NULL,
	"planned_duration" INTEGER DEFAULT 0,
	"actual_duration" INTEGER DEFAULT 0,
	"category" TEXT NOT NULL,
	"state" INTEGER DEFAULT 1,
	PRIMARY KEY("id")
);`

type dbRepo struct {
	db *sql.DB
	sync.RWMutex
}

func NewSQLiteRepo(dbfile string) (*dbRepo, error) {
	db, err := sql.Open("sqlite3", dbfile)
	if err != nil {
		return nil, err
	}

	db.SetConnMaxIdleTime(30 * time.Minute)
	db.SetMaxOpenConns(1)

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if _, err := db.Exec(createTableInterval); err != nil {
		return nil, err
	}

	return &dbRepo{
		db: db,
	}, nil
}

func (r *dbRepo) Create(i pomodoro.Interval) (int64, error) {
	r.Lock()
	defer r.Unlock()

	query := "INSERT INTO interval VALUES(NULL, ?, ?, ?, ?, ?)"
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return 0, err
	}
	defer stmt.Close()

	args := []any{
		i.StartTime,
		i.PlannedDuration,
		i.ActualDuration,
		i.Category,
		i.State,
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		return 0, err
	}

	id, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (r *dbRepo) Update(i pomodoro.Interval) error {
	r.Lock()
	defer r.Unlock()

	query := `
	UPDATE interval SET start_time=?, actual_duration=?, state=?
	WHERE id=?`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	args := []any{
		i.StartTime,
		i.ActualDuration,
		i.State,
		i.ID,
	}
	res, err := stmt.Exec(args...)
	if err != nil {
		return err
	}

	_, err = res.RowsAffected()

	return err
}

func (r *dbRepo) ByID(id int64) (pomodoro.Interval, error) {
	r.RLock()
	defer r.RUnlock()

	query := "SELECT * FROM interval WHERE id=?"

	var i pomodoro.Interval
	err := r.db.QueryRow(query, id).Scan(
		&i.ID,
		&i.StartTime,
		&i.PlannedDuration,
		&i.ActualDuration,
		&i.Category,
		&i.State,
	)

	return i, err
}

func (r *dbRepo) Last() (pomodoro.Interval, error) {
	r.RLock()
	defer r.RUnlock()

	var i pomodoro.Interval

	query := "SELECT * FROM interval ORDER BY id desc LIMIT 1"
	err := r.db.QueryRow(query).Scan(
		&i.ID,
		&i.StartTime,
		&i.PlannedDuration,
		&i.ActualDuration,
		&i.Category,
		&i.State,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return i, pomodoro.ErrNoInterval
	}
	if err != nil {
		return i, err
	}

	return i, nil
}

func (r *dbRepo) Breaks(n int) ([]pomodoro.Interval, error) {
	r.RLock()
	defer r.RUnlock()

	query := `
	SELECT * FROM interval
	WHERE	category LIKE '%Break'
	ORDER BY id DESC LIMIT ?`

	var intervals []pomodoro.Interval

	rows, err := r.db.Query(query, n)
	if err != nil {
		return nil, err
	}

	for rows.Next() {
		var i pomodoro.Interval
		err := rows.Scan(
			&i.ID,
			&i.StartTime,
			&i.PlannedDuration,
			&i.ActualDuration,
			&i.Category,
			&i.State,
		)
		if err != nil {
			return nil, err
		}

		intervals = append(intervals, i)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return intervals, nil
}

func (r *dbRepo) CategorySummary(day time.Time, filter string) (time.Duration, error) {
	r.RLock()
	defer r.RUnlock()

	query := `
	SELECT sum(actual_duration) FROM interval 
	WHERE category LIKE ? AND
	strftime('%Y-%m-%d', start_time, 'localtime')=
	strftime('%Y-%m-%d', ?, 'localtime')`

	var ds sql.NullInt64
	err := r.db.QueryRow(query, filter).Scan(&ds)

	var d time.Duration
	if ds.Valid {
		d = time.Duration(ds.Int64)
	}

	return d, err
}
