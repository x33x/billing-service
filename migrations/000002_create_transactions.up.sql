create table transactions (
    id text primary key,
    account_id text not null references accounts(id),
    amount bigint not null,
    type text not null,
    status text not null default 'pending',
    created_at timestamptz not null default now()
);