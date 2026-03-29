alter table transactions
add idempotency_key text;
create unique index uidx_idempotency_key
on transactions (idempotency_key)
where idempotency_key is not null;