package postgres

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kalin-roman/Bazar/internal/user"
)

type UserRepository struct {
	ConnectionPool *pgxpool.Pool
}

var _ user.Repository = (*UserRepository)(nil)

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{ConnectionPool: pool}
}

func (r *UserRepository) List(ctx context.Context) ([]user.User, error) {
	rows, err := r.ConnectionPool.Query(ctx,
		"select id, full_name, email, age, address_delivery, avatar from users")
	if err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()

	var users []user.User
	for rows.Next() {
		var u user.User
		if err := rows.Scan(&u.ID, &u.FullName, &u.Email, &u.Age, &u.AddressDelivery, &u.Avatar); err != nil {
			return nil, fmt.Errorf("list users: %w", err)
		}
		users = append(users, u)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("list users: %w", err)
	}

	return users, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int64) (user.User, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"select id, full_name, email, age, address_delivery, avatar from users where id = $1",
		id)

	var u user.User
	if err := row.Scan(&u.ID, &u.FullName, &u.Email, &u.Age, &u.AddressDelivery, &u.Avatar); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, fmt.Errorf("get user by id: %w", err)
	}

	return u, nil
}

func (r *UserRepository) Create(ctx context.Context, u user.User) (user.User, error) {
	row := r.ConnectionPool.QueryRow(ctx,
		"insert into users (full_name, email, age, address_delivery, avatar) values ($1, $2, $3, $4, $5) returning id",
		u.FullName, u.Email, u.Age, u.AddressDelivery, u.Avatar,
	)

	if err := row.Scan(&u.ID); err != nil {
		return user.User{}, fmt.Errorf("create user: %w", err)
	}

	return u, nil
}
