create table accounts (
    id text primary key,
    currency text not null,
    balance bigint not null default 0,
    status text not null default 'active',
    created_at timestamptz not null default now()
);