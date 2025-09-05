package subscriptions

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
)

type Repository struct {
	DB *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{DB: db}
}

var ErrNotFound = errors.New("not found")

func (r *Repository) Create(ctx context.Context, userID uuid.UUID, service string, price int, start time.Time, end *time.Time) (Subscription, error) {
	const q = `
INSERT INTO subscriptions (user_id, service_name, price, start_date, end_date)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, service_name, price, start_date, end_date, created_at, updated_at;
`
	var s Subscription
	err := r.DB.QueryRowContext(ctx, q, userID, service, price, start, end).
		Scan(&s.ID, &s.UserID, &s.ServiceName, &s.Price, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt)
	return s, err
}

func (r *Repository) GetByID(ctx context.Context, id uuid.UUID) (Subscription, error) {
	const q = `
SELECT id, user_id, service_name, price, start_date, end_date, created_at, updated_at
FROM subscriptions
WHERE id = $1;
`
	var s Subscription
	err := r.DB.QueryRowContext(ctx, q, id).
		Scan(&s.ID, &s.UserID, &s.ServiceName, &s.Price, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return Subscription{}, ErrNotFound
	}
	return s, err
}

func (r *Repository) DeleteByID(ctx context.Context, id uuid.UUID) error {
	const q = `DELETE FROM subscriptions WHERE id = $1;`
	res, err := r.DB.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *Repository) UpdateByID(ctx context.Context, id uuid.UUID, userID uuid.UUID, service string, price int, start time.Time, end *time.Time) (Subscription, error) {
	const q = `
UPDATE subscriptions
SET user_id=$2, service_name=$3, price=$4, start_date=$5, end_date=$6
WHERE id=$1
RETURNING id, user_id, service_name, price, start_date, end_date, created_at, updated_at;
`
	var s Subscription
	err := r.DB.QueryRowContext(ctx, q, id, userID, service, price, start, end).
		Scan(&s.ID, &s.UserID, &s.ServiceName, &s.Price, &s.StartDate, &s.EndDate, &s.CreatedAt, &s.UpdatedAt)
	if err == sql.ErrNoRows {
		return Subscription{}, ErrNotFound
	}
	return s, err
}
func (r *Repository) List(ctx context.Context, userID *uuid.UUID, serviceName *string, limit, offset int) ([]Subscription, error) {
	q := `
SELECT id, user_id, service_name, price, start_date, end_date, created_at, updated_at
FROM subscriptions
`
	var where []string
	var args []any

	if userID != nil {
		where = append(where, fmt.Sprintf("user_id = $%d", len(args)+1))
		args = append(args, *userID)
	}

	if serviceName != nil && *serviceName != "" {
		where = append(where, fmt.Sprintf("service_name = $%d", len(args)+1))
		args = append(args, *serviceName)
	}

	if len(where) > 0 {
		q += "WHERE " + strings.Join(where, " AND ") + "\n"
	}

	q += fmt.Sprintf("ORDER BY created_at DESC LIMIT $%d OFFSET $%d", len(args)+1, len(args)+2)
	args = append(args, limit, offset)

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Subscription
	for rows.Next() {
		var s Subscription
		if err := rows.Scan(
			&s.ID,
			&s.UserID,
			&s.ServiceName,
			&s.Price,
			&s.StartDate,
			&s.EndDate,
			&s.CreatedAt,
			&s.UpdatedAt,
		); err != nil {
			return nil, err
		}
		list = append(list, s)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return list, nil
}

func (r *Repository) Summary(ctx context.Context, from, to time.Time, userID *uuid.UUID, serviceName *string) (int, error) {
	const base = `
SELECT price, start_date, end_date
FROM subscriptions
WHERE start_date <= $1
  AND (end_date IS NULL OR end_date >= $2)
`
	args := []any{to, from}
	where := ""

	if userID != nil {
		where += " AND user_id = $" + strconv.Itoa(len(args)+1)
		args = append(args, *userID)
	}
	if serviceName != nil && *serviceName != "" {
		where += " AND service_name = $" + strconv.Itoa(len(args)+1)
		args = append(args, *serviceName)
	}

	q := base + where

	rows, err := r.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return 0, err
	}
	defer rows.Close()

	ffrom := time.Date(from.Year(), from.Month(), 1, 0, 0, 0, 0, time.UTC)
	tto := time.Date(to.Year(), to.Month(), 1, 0, 0, 0, 0, time.UTC)

	total := 0
	for rows.Next() {
		var price int
		var s time.Time
		var e *time.Time
		if err := rows.Scan(&price, &s, &e); err != nil {
			return 0, err
		}

		s = time.Date(s.Year(), s.Month(), 1, 0, 0, 0, 0, time.UTC)
		var end time.Time
		if e == nil {
			end = tto
		} else {
			end = time.Date(e.Year(), e.Month(), 1, 0, 0, 0, 0, time.UTC)
		}

		aStart := maxDate(s, ffrom)
		aEnd := minDate(end, tto)
		if aEnd.Before(aStart) {
			continue
		}
		months := (aEnd.Year()-aStart.Year())*12 + int(aEnd.Month()) - int(aStart.Month()) + 1
		total += months * price
	}
	if err := rows.Err(); err != nil {
		return 0, err
	}
	return total, nil
}

func minDate(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func maxDate(a, b time.Time) time.Time {
	if a.After(b) {
		return a
	}
	return b
}
