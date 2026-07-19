create table if not exists users(
    id bigserial primary key,
    login text  not null unique,
    password_hash text not null,
    role text not null default 'bidder'
    check (role in ('bidder', 'seller', 'admin')),
created_at timestamp not null default now()
);