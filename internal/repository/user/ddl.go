package user

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const ddlUser = `create table users (
    id bigserial primary key,
    login text  not null unique,
    password_hash text not null,
    role text not null default 'bidder'
        check (role in ('bidder', 'seller', 'admin')),
    created_at timestamp not null default now()
)`

func runDdl(ctx context.Context, pool *pgxpool.Pool) error {
	_, err := pool.Exec(ctx, ddlUser)
	if err != nil {
		return err
	}
	return nil
}
