-- name: GetSAMLConnectionByID :one
select *
from saml_connections
where id = $1;

-- name: GetOrganizationByID :one
select *
from organizations
where id = $1;

-- name: GetEnvironmentByID :one
select *
from environments
where id = $1;

-- name: CreateSAMLSession :one
insert into saml_sessions (id, saml_connection_id, secret_access_token, subject_id, subject_idp_attributes)
values ($1, $2, $3, $4, $5)
returning *;

-- name: GetAPIKeyBySecretValue :one
select *
from api_keys
where secret_value = $1;

-- name: GetSAMLAccessTokenData :one
select saml_sessions.*, organizations.id as organization_id, environments.id as environment_id
from saml_sessions
         join saml_connections on saml_sessions.saml_connection_id = saml_connections.id
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_sessions.secret_access_token = $2;

-- name: GetAppUserByEmail :one
select *
from app_users
where email = $1;

-- name: GetAppUserByID :one
select *
from app_users
where app_organization_id = $1
  and id = $2;

-- name: GetAppOrganizationByGoogleHostedDomain :one
select *
from app_organizations
where google_hosted_domain = $1;

-- name: CreateAppOrganization :one
insert into app_organizations (id, google_hosted_domain)
values ($1, $2)
returning *;

-- name: CreateAppUser :one
insert into app_users (id, app_organization_id, display_name, email)
values ($1, $2, $3, $4)
returning *;

-- name: CreateAppSession :one
insert into app_sessions (id, app_user_id, create_time, expire_time, token)
values ($1, $2, $3, $4, $5)
returning *;

-- name: GetAppSessionByToken :one
select app_sessions.app_user_id, app_users.app_organization_id
from app_sessions
         join app_users on app_sessions.app_user_id = app_users.id
where token = $1
  and expire_time > $2;
