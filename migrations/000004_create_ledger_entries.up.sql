create table ledger_entries (
    id text primary key,
    transaction_id text not null REFERENCES transactions(id),
    account_id text not null REFERENCES accounts(id),
    amount bigint not null default 0,
    direction text not null,
    created_at timestamptz not null default now()
);