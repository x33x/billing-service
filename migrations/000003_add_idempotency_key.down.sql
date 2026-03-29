drop index uidx_idempotency_key;
alter table transactions drop column idempotency_key;