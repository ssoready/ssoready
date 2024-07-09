alter table app_sessions add column token_sha256 bytea unique;
alter table app_sessions drop constraint app_sessions_token_key;
