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

-- name: GetSAMLSessionBySecretAccessToken :one
select saml_sessions.*
from saml_sessions
         join saml_connections on saml_sessions.saml_connection_id = saml_connections.id
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_sessions.secret_access_token = $2;
