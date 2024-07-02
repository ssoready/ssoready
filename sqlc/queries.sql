-- name: CheckExistsEmailVerificationChallenge :one
select exists(select * from email_verification_challenges where email = $1 and expire_time > $2);

-- name: CreateEmailVerificationChallenge :one
insert into email_verification_challenges (id, email, expire_time, secret_token)
values ($1, $2, $3, $4)
returning *;

-- name: GetEmailVerificationChallengeBySecretToken :one
select *
from email_verification_challenges
where secret_token = $1
  and expire_time > $2;

-- name: GetOnboardingState :one
select *
from onboarding_states
where app_organization_id = $1;

-- name: UpdateOnboardingState :one
insert into onboarding_states (app_organization_id, dummyidp_app_id, onboarding_environment_id,
                               onboarding_organization_id,
                               onboarding_saml_connection_id)
values ($1, $2, $3, $4, $5)
on conflict (app_organization_id) do update set dummyidp_app_id               = excluded.dummyidp_app_id,
                                                onboarding_environment_id     = excluded.onboarding_environment_id,
                                                onboarding_organization_id    = excluded.onboarding_organization_id,
                                                onboarding_saml_connection_id = excluded.onboarding_saml_connection_id
returning *;

-- name: AuthGetInitData :one
select idp_redirect_url, sp_entity_id
from saml_connections
where saml_connections.id = $1;

-- name: AuthGetValidateData :one
select saml_connections.sp_entity_id,
       saml_connections.idp_entity_id,
       saml_connections.idp_x509_certificate,
       environments.redirect_url,
       environments.oauth_redirect_uri
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where saml_connections.id = $1;

-- name: AuthCheckAssertionAlreadyProcessed :one
select exists(select * from saml_flows where id = $1 and access_code_sha256 is not null);

-- name: AuthGetSAMLConnectionDomains :many
select organization_domains.domain
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
         join organization_domains on organizations.id = organization_domains.organization_id
where saml_connections.id = $1;

-- name: CreateSAMLFlowGetRedirect :one
insert into saml_flows (id, saml_connection_id, expire_time, state, create_time, update_time,
                        auth_redirect_url, get_redirect_time, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9)
returning *;

-- name: UpsertSAMLFlowInitiate :one
insert into saml_flows (id, saml_connection_id, expire_time, state, create_time, update_time,
                        initiate_request, initiate_time, status, is_oauth)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
on conflict (id) do update set update_time      = excluded.update_time,
                               initiate_request = excluded.initiate_request,
                               initiate_time    = excluded.initiate_time,
                               status           = excluded.status
returning *;

-- name: UpsertSAMLFlowReceiveAssertion :one
insert into saml_flows (id, saml_connection_id, access_code_sha256, expire_time, state, create_time, update_time,
                        assertion, receive_assertion_time, error_unsigned_assertion, error_bad_issuer,
                        error_bad_audience, error_bad_subject_id, error_email_outside_organization_domains, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
on conflict (id) do update set access_code_sha256                       = excluded.access_code_sha256,
                               update_time                              = excluded.update_time,
                               assertion                                = excluded.assertion,
                               receive_assertion_time                   = excluded.receive_assertion_time,
                               error_unsigned_assertion                 = excluded.error_unsigned_assertion,
                               error_bad_issuer                         = excluded.error_bad_issuer,
                               error_bad_audience                       = excluded.error_bad_audience,
                               error_bad_subject_id                     = excluded.error_bad_subject_id,
                               error_email_outside_organization_domains = excluded.error_email_outside_organization_domains,
                               status                                   = excluded.status
returning *;

-- name: UpdateSAMLFlowRedeem :one
update saml_flows
set update_time        = $1,
    redeem_time        = $2,
    redeem_response    = $3,
    status             = $4,
    access_code_sha256 = null
where id = $5
returning *;

-- name: AuthGetSAMLFlow :one
select *
from saml_flows
where id = $1;

-- name: UpdateSAMLFlowSubjectData :one
update saml_flows
set email                  = $1,
    subject_idp_attributes = $2
where id = $3
returning *;

-- name: GetPrimarySAMLConnectionIDByOrganizationID :one
select saml_connections.id
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.id = $2
  and saml_connections.is_primary = true;

-- name: GetPrimarySAMLConnectionIDByOrganizationExternalID :one
select saml_connections.id
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.external_id = $2
  and saml_connections.is_primary = true;

-- name: GetSAMLRedirectURLData :one
select environments.auth_url
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and environments.id = @environment_id
  and saml_connections.id = @saml_connection_id;

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

-- name: GetAPIKeyBySecretValueSHA256 :one
select api_keys.*, environments.app_organization_id
from api_keys
         join environments on api_keys.environment_id = environments.id
where secret_value_sha256 = $1;

-- name: GetSAMLAccessCodeData :one
select saml_flows.id             as saml_flow_id,
       saml_flows.email,
       saml_flows.subject_idp_attributes,
       saml_flows.state,
       organizations.id          as organization_id,
       organizations.external_id as organization_external_id,
       environments.id           as environment_id
from saml_flows
         join saml_connections on saml_flows.saml_connection_id = saml_connections.id
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and environments.id = @environment_id
  and saml_flows.access_code_sha256 = $2;

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
insert into app_sessions (id, app_user_id, create_time, expire_time, token, token_sha256, revoked)
values ($1, $2, $3, $4, '', $5, $6)
returning *;

-- name: RevokeAppSessionByID :one
update app_sessions
set revoked = true
where id = $1
returning *;

-- name: GetAppSessionByTokenSHA256 :one
select app_sessions.id,
       app_sessions.app_user_id,
       app_users.display_name,
       app_users.email,
       app_users.app_organization_id
from app_sessions
         join app_users on app_sessions.app_user_id = app_users.id
where token_sha256 = $1
  and expire_time > $2
  and revoked = false;

-- name: ListEnvironments :many
select *
from environments
where app_organization_id = $1
  and id > $2
order by id
limit $3;

-- name: GetEnvironment :one
select *
from environments
where app_organization_id = $1
  and id = $2;

-- name: CreateEnvironment :one
insert into environments (id, redirect_url, app_organization_id, display_name, auth_url)
values ($1, $2, $3, $4, $5)
returning *;

-- name: UpdateEnvironment :one
update environments
set display_name       = $1,
    redirect_url       = $2,
    auth_url           = $3,
    oauth_redirect_uri = $4
where id = $5
returning *;

-- name: ListAPIKeys :many
select *
from api_keys
where environment_id = $1
  and id > $2
order by id
limit $3;

-- name: GetAPIKey :one
select api_keys.*
from api_keys
         join environments on api_keys.environment_id = environments.id
where environments.app_organization_id = $1
  and api_keys.id = $2;

-- name: CreateAPIKey :one
insert into api_keys (id, secret_value, secret_value_sha256, environment_id)
values ($1, '', $2, $3)
returning *;

-- name: DeleteAPIKey :exec
delete
from api_keys
where id = $1;

-- name: ListOrganizations :many
select *
from organizations
where environment_id = $1
  and id >= $2
order by id
limit $3;

-- name: GetOrganization :one
select organizations.*
from organizations
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and organizations.id = $2;

-- name: ListOrganizationDomains :many
select *
from organization_domains
where organization_id = any ($1::uuid[]);

-- name: CreateOrganization :one
insert into organizations (id, environment_id, external_id)
values ($1, $2, $3)
returning *;

-- name: CreateOrganizationDomain :one
insert into organization_domains (id, organization_id, domain)
values ($1, $2, $3)
returning *;

-- name: UpdateOrganization :one
update organizations
set external_id = $1
where id = $2
returning *;

-- name: DeleteOrganizationDomains :exec
delete
from organization_domains
where organization_id = $1;

-- name: ListSAMLConnections :many
select *
from saml_connections
where organization_id = $1
  and id >= $2
order by id
limit $3;

-- name: GetSAMLConnection :one
select saml_connections.*
from saml_connections
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_connections.id = $2;

-- name: CreateSAMLConnection :one
insert into saml_connections (id, organization_id, sp_entity_id, sp_acs_url, idp_entity_id, idp_redirect_url,
                              idp_x509_certificate,
                              is_primary)
values ($1, $2, $3, $4, $5, $6, $7, $8)
returning *;

-- name: UpdateSAMLConnection :one
update saml_connections
set idp_entity_id        = $1,
    idp_redirect_url     = $2,
    idp_x509_certificate = $3,
    is_primary           = $4
where id = $5
returning *;

-- name: UpdatePrimarySAMLConnection :exec
update saml_connections
set is_primary = (id = $1)
where organization_id = $2;

-- name: ListSAMLFlowsFirstPage :many
select *
from saml_flows
where saml_connection_id = $1
order by (create_time, id) desc
limit $2;

-- name: ListSAMLFlowsNextPage :many
select *
from saml_flows
where saml_connection_id = $1
  and (create_time, id) <= (@create_time, @id::uuid)
order by (create_time, id) desc
limit $2;

-- name: GetSAMLFlow :one
select saml_flows.*
from saml_flows
         join saml_connections on saml_flows.saml_connection_id = saml_connections.id
         join organizations on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_flows.id = $2;

-- name: ListSAMLOAuthClients :many
select *
from saml_oauth_clients
where environment_id = $1
  and id > $2
order by id
limit $3;

-- name: GetSAMLOAuthClient :one
select saml_oauth_clients.*
from saml_oauth_clients
         join environments on saml_oauth_clients.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_oauth_clients.id = $2;

-- name: CreateSAMLOAuthClient :one
insert into saml_oauth_clients (id, environment_id, client_secret_sha256)
values ($1, $2, $3)
returning *;

-- name: DeleteSAMLOAuthClient :exec
delete
from saml_oauth_clients
where id = $1;

-- name: AuthGetSAMLOAuthClient :one
select saml_oauth_clients.*, environments.app_organization_id
from saml_oauth_clients
         join environments on saml_oauth_clients.environment_id = environments.id
where saml_oauth_clients.id = $1;

-- name: AuthGetSAMLOAuthClientWithSecret :one
select saml_oauth_clients.*, environments.app_organization_id
from saml_oauth_clients
         join environments on saml_oauth_clients.environment_id = environments.id
where saml_oauth_clients.id = $1
  and saml_oauth_clients.client_secret_sha256 = $2;
