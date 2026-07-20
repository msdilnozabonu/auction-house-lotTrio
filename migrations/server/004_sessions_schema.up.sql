create table if not exists sessions
(
    id BIGINT generated always as identity primary key,
    user_id BIGINT not null references users(id) on delete cascade,
    refresh_token TEXT not null unique,
    created_at timestamptz not null default now(),
    updated_at    timestamptz  not null default now(),
    expires_at timestamptz not null default now() + interval '10 days'
);