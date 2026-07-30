create table reports(
    id bigserial primary key,
    lot_id bigint not null references lots(id),
    reporter_id bigint not null references users(id),
    reason text not null,
    status text not null default 'pending' check (status in ('pending', 'resolved', 'dismissed')),
    created_at timestamp not null default now()
);