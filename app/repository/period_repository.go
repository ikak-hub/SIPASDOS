package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"sipasdos/app/model"
)

type PeriodRepository interface {
	FindAll(ctx context.Context, q model.ListQuery) ([]model.RecruitmentPeriod, int, error)
	FindByID(ctx context.Context, id int) (model.RecruitmentPeriod, error)
	Create(ctx context.Context, p model.RecruitmentPeriod) (model.RecruitmentPeriod, error)
	Update(ctx context.Context, p model.RecruitmentPeriod) (model.RecruitmentPeriod, error)
	Delete(ctx context.Context, id int) error
}

type periodRepository struct {
	pool *pgxpool.Pool
}

func NewPeriodRepository(pool *pgxpool.Pool) PeriodRepository {
	return &periodRepository{pool: pool}
}

const periodColumns = "id, name, start_date, end_date, status, created_at"

var periodSortColumns = map[string]string{
	"id":         "id",
	"name":       "name",
	"start_date": "start_date",
	"end_date":   "end_date",
	"created_at": "created_at",
}

func scanPeriod(row pgx.Row) (model.RecruitmentPeriod, error) {
	var p model.RecruitmentPeriod
	err := row.Scan(&p.ID, &p.Name, &p.StartDate, &p.EndDate, &p.Status, &p.CreatedAt)
	return p, err
}

func (r *periodRepository) FindAll(
	ctx context.Context, q model.ListQuery,
) ([]model.RecruitmentPeriod, int, error) {
	where := "WHERE 1=1"
	args := []any{}

	if q.Search != "" {
		args = append(args, "%"+q.Search+"%")
		where += fmt.Sprintf(" AND name ILIKE $%d", len(args))
	}
	if q.Status != "" {
		args = append(args, q.Status)
		where += fmt.Sprintf(" AND status = $%d", len(args))
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM recruitment_periods " + where
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	column, ok := periodSortColumns[q.Sort]
	if !ok {
		column = "id"
	}
	direction := "ASC"
	if q.Order == "desc" {
		direction = "DESC"
	}

	args = append(args, q.Limit, (q.Page-1)*q.Limit)
	query := fmt.Sprintf(
		"SELECT %s FROM recruitment_periods %s ORDER BY %s %s LIMIT $%d OFFSET $%d",
		periodColumns, where, column, direction, len(args)-1, len(args),
	)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	periods := []model.RecruitmentPeriod{}
	for rows.Next() {
		p, err := scanPeriod(rows)
		if err != nil {
			return nil, 0, err
		}
		periods = append(periods, p)
	}
	return periods, total, rows.Err()
}

func (r *periodRepository) FindByID(ctx context.Context, id int) (model.RecruitmentPeriod, error) {
	p, err := scanPeriod(r.pool.QueryRow(ctx,
		"SELECT "+periodColumns+" FROM recruitment_periods WHERE id = $1", id))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RecruitmentPeriod{}, ErrNotFound
		}
		return model.RecruitmentPeriod{}, err
	}
	return p, nil
}

func (r *periodRepository) Create(
	ctx context.Context, p model.RecruitmentPeriod,
) (model.RecruitmentPeriod, error) {
	created, err := scanPeriod(r.pool.QueryRow(ctx,
		`INSERT INTO recruitment_periods (name, start_date, end_date, status)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+periodColumns,
		p.Name, p.StartDate, p.EndDate, p.Status))
	if err != nil {
		return model.RecruitmentPeriod{}, mapWriteError(err)
	}
	return created, nil
}

func (r *periodRepository) Update(
	ctx context.Context, p model.RecruitmentPeriod,
) (model.RecruitmentPeriod, error) {
	updated, err := scanPeriod(r.pool.QueryRow(ctx,
		`UPDATE recruitment_periods
		 SET name = $2, start_date = $3, end_date = $4, status = $5
		 WHERE id = $1
		 RETURNING `+periodColumns,
		p.ID, p.Name, p.StartDate, p.EndDate, p.Status))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.RecruitmentPeriod{}, ErrNotFound
		}
		return model.RecruitmentPeriod{}, mapWriteError(err)
	}
	return updated, nil
}

func (r *periodRepository) Delete(ctx context.Context, id int) error {
	tag, err := r.pool.Exec(ctx, "DELETE FROM recruitment_periods WHERE id = $1", id)
	if err != nil {
		if pgErrorCode(err) == pgForeignKeyViolation {
			return ErrInUse // masih punya lowongan
		}
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
