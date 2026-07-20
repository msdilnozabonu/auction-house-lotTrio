create table if not exists lots (
                                    id bigserial primary key,
                                    seller_id bigint not null references users(id),
    title text not null,
    description text,
    start_price numeric not null check ( start_price > 0 ),
    current_price numeric not null check ( current_price > 0 ),
    current_winner_id bigint references users(id),
    status text not null default 'draft'
    check (status in ('draft', 'active', 'closed', 'cancelled')),
    starts_at timestamp not null default now() + interval '1 day',
    ends_at timestamp,
    photo_path text,
    created_at timestamp not null default now()
    );