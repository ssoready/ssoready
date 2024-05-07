create table email_verification_challenges
(
    id           uuid        not null primary key,
    email        varchar     not null,
    expire_time  timestamptz not null,
    secret_token varchar     not null
);
