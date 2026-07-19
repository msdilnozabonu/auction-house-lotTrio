create table if not exists bids (
                                    id bigserial primary key,
                                    lot_id bigint not null references lots(id),
    bidder_id bigint not null references users(id),
    amount numeric not null check (amount > 0),
    created_at timestamp not null default now()
    );
