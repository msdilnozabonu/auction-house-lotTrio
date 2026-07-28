create table if not exists watchlist (
    user_id BIGINT not null references users(id) on delete cascade,
    lot_id BIGINT not null references lots(id) on delete cascade,
    created_at timestamp not null default now(),
    primary key (user_id, lot_id)
    );