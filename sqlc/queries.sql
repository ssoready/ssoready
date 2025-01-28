-- name: CheckExistsEmailVerificationChallenge :one
select exists(select *
              from email_verification_challenges
              where email = $1
                and expire_time > $2
                and complete_time is null);

-- name: CreateEmailVerificationChallenge :one
insert into email_verification_challenges (id, email, expire_time, secret_token)
values ($1, $2, $3, $4)
returning *;

-- name: GetEmailVerificationChallengeBySecretToken :one
select *
from email_verification_challenges
where secret_token = $1
  and expire_time > $2;

-- name: UpdateEmailVerificationChallengeCompleteTime :one
update email_verification_challenges
set complete_time = $1
where id = $2
returning *;

-- name: GetOnboardingState :one
select *
from onboarding_states
where app_organization_id = $1;

-- name: UpdateOnboardingState :one
insert into onboarding_states (app_organization_id, dummyidp_app_id,
                               onboarding_environment_id,
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
       environments.oauth_redirect_uri,
       environments.admin_url
from saml_connections
         join organizations
              on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where saml_connections.id = $1;

-- name: AuthCheckAssertionAlreadyProcessed :one
select exists(select *
              from saml_flows
              where id = $1
                and access_code_sha256 is not null);

-- name: AuthGetSAMLConnectionDomains :many
select organization_domains.domain
from saml_connections
         join organizations
              on saml_connections.organization_id = organizations.id
         join organization_domains
              on organizations.id = organization_domains.organization_id
where saml_connections.id = $1;

-- name: CreateSAMLFlowGetRedirect :one
insert into saml_flows (id, saml_connection_id, expire_time, state, create_time,
                        update_time,
                        auth_redirect_url, get_redirect_time, status,
                        test_mode_idp,
                        error_saml_connection_not_configured,
                        error_environment_oauth_redirect_uri_not_configured)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
returning *;

-- name: UpsertSAMLFlowInitiate :one
insert into saml_flows (id, saml_connection_id, expire_time, state, create_time,
                        update_time,
                        initiate_request, initiate_time, status, is_oauth)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
on conflict (id) do update set update_time      = excluded.update_time,
                               initiate_request = excluded.initiate_request,
                               initiate_time    = excluded.initiate_time,
                               status           = excluded.status
returning *;

-- name: UpsertSAMLFlowReceiveAssertion :one
insert into saml_flows (id, saml_connection_id, assertion_id,
                        access_code_sha256, expire_time, state, create_time,
                        update_time, assertion, receive_assertion_time,
                        error_saml_connection_not_configured,
                        error_unsigned_assertion, error_bad_issuer,
                        error_bad_audience, error_bad_signature_algorithm,
                        error_bad_digest_algorithm, error_bad_x509_certificate,
                        error_bad_subject_id,
                        error_email_outside_organization_domains, status)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15, $16,
        $17, $18, $19, $20)
on conflict (id) do update set assertion_id                             = excluded.assertion_id,
                               access_code_sha256                       = excluded.access_code_sha256,
                               update_time                              = excluded.update_time,
                               assertion                                = excluded.assertion,
                               receive_assertion_time                   = excluded.receive_assertion_time,
                               error_saml_connection_not_configured     = excluded.error_saml_connection_not_configured,
                               error_unsigned_assertion                 = excluded.error_unsigned_assertion,
                               error_bad_issuer                         = excluded.error_bad_issuer,
                               error_bad_audience                       = excluded.error_bad_audience,
                               error_bad_signature_algorithm            = excluded.error_bad_signature_algorithm,
                               error_bad_digest_algorithm               = excluded.error_bad_digest_algorithm,
                               error_bad_x509_certificate               = excluded.error_bad_x509_certificate,
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
         join organizations
              on saml_connections.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.id = $2
  and saml_connections.is_primary = true;

-- name: GetPrimarySAMLConnectionIDByOrganizationExternalID :one
select saml_connections.id
from saml_connections
         join organizations
              on saml_connections.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.external_id = $2
  and saml_connections.is_primary = true;

-- name: GetSAMLRedirectURLData :one
select environments.auth_url as environment_auth_url,
       saml_connections.idp_entity_id,
       saml_connections.idp_redirect_url,
       saml_connections.idp_x509_certificate
from saml_connections
         join organizations
              on saml_connections.organization_id = organizations.id
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
         join saml_connections
              on saml_flows.saml_connection_id = saml_connections.id
         join organizations
              on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and environments.id = @environment_id
  and saml_flows.access_code_sha256 = $2;

-- name: GetAppUserByEmail :one
select *
from app_users
where email = $1;

-- name: ListAppUsers :many
select id, display_name, email
from app_users
where app_organization_id = $1;

-- name: GetAppUserByID :one
select *
from app_users
where app_organization_id = $1
  and id = $2;

-- name: GetAppOrganizationByID :one
select *
from app_organizations
where id = $1;

-- name: GetAppOrganizationByGoogleHostedDomain :one
select *
from app_organizations
where google_hosted_domain = $1;

-- name: GetAppOrganizationByMicrosoftTenantID :one
select *
from app_organizations
where microsoft_tenant_id = $1;

-- name: CreateAppOrganization :one
insert into app_organizations (id, google_hosted_domain, microsoft_tenant_id)
values ($1, $2, $3)
returning *;

-- name: CreateAppUser :one
insert into app_users (id, app_organization_id, display_name, email)
values ($1, $2, $3, $4)
returning *;

-- name: CreateAppSession :one
insert into app_sessions (id, app_user_id, create_time, expire_time, token,
                          token_sha256, revoked)
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
  and id >= $2
order by id
limit $3;

-- name: GetEnvironment :one
select *
from environments
where app_organization_id = $1
  and id = $2;

-- name: CreateEnvironment :one
insert into environments (id, redirect_url, oauth_redirect_uri,
                          app_organization_id, display_name, auth_url)
values ($1, $2, $3, $4, $5, $6)
returning *;

-- name: UpdateEnvironment :one
update environments
set display_name       = $1,
    redirect_url       = $2,
    auth_url           = $3,
    oauth_redirect_uri = $4
where id = $5
returning *;

-- name: UpdateEnvironmentCustomAuthDomain :one
update environments
set custom_auth_domain = $1
where id = $2
returning *;

-- name: UpdateEnvironmentCustomAdminDomain :one
update environments
set custom_admin_domain = $1
where id = $2
returning *;

-- name: UpdateEnvironmentAuthURL :one
update environments
set auth_url = $1
where id = $2
returning *;

-- name: UpdateEnvironmentAdminURL :one
update environments
set admin_url = $1
where id = $2
returning *;

-- name: ListAPIKeys :many
select *
from api_keys
where environment_id = $1
  and id >= $2
order by id
limit $3;

-- name: GetAPIKey :one
select api_keys.*
from api_keys
         join environments on api_keys.environment_id = environments.id
where environments.app_organization_id = $1
  and api_keys.id = $2;

-- name: CreateAPIKey :one
insert into api_keys (id, secret_value, secret_value_sha256, environment_id,
                      has_management_api_access)
values ($1, '', $2, $3, $4)
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

-- name: ManagementGetOrganization :one
select *
from organizations
where environment_id = $1
  and id = $2;

-- name: ListOrganizationDomains :many
select *
from organization_domains
where organization_id = any ($1::uuid[]);

-- name: CreateOrganization :one
insert into organizations (id, environment_id, external_id, display_name)
values ($1, $2, $3, $4)
returning *;

-- name: CreateOrganizationDomain :one
insert into organization_domains (id, organization_id, domain)
values ($1, $2, $3)
returning *;

-- name: UpdateOrganization :one
update organizations
set external_id  = $1,
    display_name = $2
where id = $3
returning *;

-- name: DeleteOrganization :execrows
delete
from organizations
where id = $1;

-- name: DeleteOrganizationDomains :exec
delete
from organization_domains
where organization_id = $1;

-- name: DeleteOrganizationAdminAccessTokens :execrows
delete
from admin_access_tokens
where organization_id = $1;

-- name: ListAllSAMLConnectionIDs :many
select id
from saml_connections
where organization_id = $1;

-- name: ListAllSCIMDirectoryIDs :many
select id
from scim_directories
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
         join organizations
              on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_connections.id = $2;

-- name: ManagementGetSAMLConnection :one
select saml_connections.*
from saml_connections
         join organizations
              on saml_connections.organization_id = organizations.id
where organizations.environment_id = $1
  and saml_connections.id = $2;

-- name: CreateSAMLConnection :one
insert into saml_connections (id, organization_id, sp_entity_id, sp_acs_url,
                              idp_entity_id, idp_redirect_url,
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

-- name: DeleteSAMLConnection :execrows
delete
from saml_connections
where id = $1;

-- name: DeleteSAMLFlowsBySAMLConnectionID :execrows
delete
from saml_flows
where saml_connection_id = $1;

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
         join saml_connections
              on saml_flows.saml_connection_id = saml_connections.id
         join organizations
              on saml_connections.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and saml_flows.id = $2;

-- name: GetSAMLFlowByID :one
select *
from saml_flows
where id = $1;

-- name: ListSAMLOAuthClients :many
select *
from saml_oauth_clients
where environment_id = $1
  and id >= $2
order by id
limit $3;

-- name: GetSAMLOAuthClient :one
select saml_oauth_clients.*
from saml_oauth_clients
         join environments
              on saml_oauth_clients.environment_id = environments.id
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
         join environments
              on saml_oauth_clients.environment_id = environments.id
where saml_oauth_clients.id = $1;

-- name: AuthGetSAMLOAuthClientWithSecret :one
select saml_oauth_clients.*, environments.app_organization_id
from saml_oauth_clients
         join environments
              on saml_oauth_clients.environment_id = environments.id
where saml_oauth_clients.id = $1
  and saml_oauth_clients.client_secret_sha256 = $2;

-- name: UpdateEnvironmentAdminSettings :one
update environments
set admin_application_name = $1,
    admin_return_url       = $2
where id = $3
returning *;

-- name: UpdateEnvironmentAdminLogoConfigured :one
update environments
set admin_logo_configured = $1
where id = $2
returning *;

-- name: CreateAdminAccessToken :one
insert into admin_access_tokens (id, organization_id, one_time_token_sha256,
                                 create_time, expire_time, can_manage_saml,
                                 can_manage_scim)
values ($1, $2, $3, $4, $5, $6, $7)
returning *;

-- name: AdminGetAdminAccessTokenByOneTimeToken :one
select *
from admin_access_tokens
where one_time_token_sha256 = $1;

-- name: AdminGetAdminAccessTokenByAccessToken :one
select *
from admin_access_tokens
where access_token_sha256 = $1
  and expire_time > $2;

-- name: AdminConvertAdminAccessTokenToSession :one
update admin_access_tokens
set one_time_token_sha256 = null,
    access_token_sha256   = $1
where id = $2
returning *;

-- name: AdminGetSAMLConnection :one
select *
from saml_connections
where organization_id = $1
  and id = $2;

-- name: AdminGetSCIMDirectory :one
select *
from scim_directories
where organization_id = $1
  and id = $2;

-- name: GetSCIMDirectoryByID :one
select *
from scim_directories
where id = $1;

-- name: AuthGetSCIMDirectory :one
select *
from scim_directories
where id = $1;

-- name: AuthGetSCIMDirectoryOrganizationDomains :many
select organization_domains.domain
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
         join organization_domains
              on organizations.id = organization_domains.organization_id
where scim_directories.id = $1;

-- name: AuthGetSCIMDirectoryByIDAndBearerToken :one
select *
from scim_directories
where id = $1
  and bearer_token_sha256 = $2;

-- name: AuthCountSCIMUsers :one
select count(*)
from scim_users
where scim_directory_id = $1
  and deleted = false;

-- name: AuthListSCIMUsers :many
select *
from scim_users
where scim_directory_id = $1
  and deleted = false
order by id
offset $2 limit $3;

-- name: AuthGetSCIMUserByEmail :one
select *
from scim_users
where scim_directory_id = $1
  and email = $2
  and deleted = false;

-- name: AuthGetSCIMUser :one
select *
from scim_users
where scim_directory_id = $1
  and id = $2
  and deleted = false;

-- name: AuthGetSCIMUserIncludeDeleted :one
select *
from scim_users
where scim_directory_id = $1
  and id = $2;

-- name: AuthUpsertSCIMUser :one
insert into scim_users (id, scim_directory_id, email, deleted, attributes)
values ($1, $2, $3, $4, $5)
on conflict (scim_directory_id, email) do update set deleted    = excluded.deleted,
                                                     attributes = excluded.attributes
returning *;

-- name: AuthUpdateSCIMUser :one
update scim_users
set email      = $1,
    attributes = $2,
    deleted    = $5
where scim_directory_id = $3
  and id = $4
returning *;

-- name: AuthUpdateSCIMUserEmail :one
update scim_users
set email = $1
where scim_directory_id = $2
  and id = $3
returning *;

-- name: AuthMarkSCIMUserDeleted :one
update scim_users
set deleted = true
where id = $1
returning *;

-- name: AuthCountSCIMGroups :one
select count(*)
from scim_groups
where scim_directory_id = $1
  and deleted = false;

-- name: AuthListSCIMGroups :many
select *
from scim_groups
where scim_directory_id = $1
  and deleted = false
order by id
offset $2 limit $3;

-- name: AuthCountSCIMGroupsByDisplayName :one
select count(*)
from scim_groups
where scim_directory_id = $1
  and deleted = false
  and display_name = $2;

-- name: AuthListSCIMGroupsByDisplayName :many
select *
from scim_groups
where scim_directory_id = $1
  and deleted = false
  and display_name = $2
order by id
offset $3 limit $4;

-- name: AuthGetSCIMGroup :one
select *
from scim_groups
where scim_directory_id = $1
  and id = $2;

-- name: AuthCreateSCIMGroup :one
insert into scim_groups (id, scim_directory_id, display_name, attributes,
                         deleted)
values ($1, $2, $3, $4, $5)
returning *;

-- name: AuthUpdateSCIMGroup :one
update scim_groups
set display_name = $1,
    attributes   = $2
where id = $3
returning *;

-- name: AuthUpdateSCIMGroupDisplayName :one
update scim_groups
set display_name = $1
where id = $2
returning *;

-- name: AuthClearSCIMGroupMembers :exec
delete
from scim_user_group_memberships
where scim_group_id = $1;

-- name: AuthMarkSCIMGroupDeleted :one
update scim_groups
set deleted = true
where id = $1
returning *;

-- name: AuthUpsertSCIMUserGroupMembership :exec
insert into scim_user_group_memberships (id, scim_directory_id, scim_user_id,
                                         scim_group_id)
values ($1, $2, $3, $4)
on conflict (scim_user_id, scim_group_id) do nothing;

-- name: AuthDeleteSCIMUserGroupMembership :exec
delete
from scim_user_group_memberships
where scim_directory_id = $1
  and scim_user_id = $2
  and scim_group_id = $3;

-- name: AuthCreateSCIMRequest :one
insert into scim_requests (id, scim_directory_id, timestamp, http_request_url,
                           http_request_method,
                           http_request_body, http_response_status,
                           http_response_body, error_bad_bearer_token,
                           error_bad_username,
                           error_email_outside_organization_domains)
values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
returning *;

-- name: AppListSCIMRequests :many
select *
from scim_requests
where scim_directory_id = $1
  and id <= $2
order by id desc
limit $3;

-- name: AppGetSCIMRequest :one
select scim_requests.*
from scim_requests
         join scim_directories
              on scim_requests.scim_directory_id = scim_directories.id
         join organizations
              on scim_directories.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and scim_requests.id = $2;

-- name: GetSCIMDirectoryByIDAndEnvironmentID :one
select scim_directories.id
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and scim_directories.id = $2;

-- name: GetPrimarySCIMDirectoryIDByOrganizationID :one
select scim_directories.id
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.id = $2
  and scim_directories.is_primary = true;

-- name: GetPrimarySCIMDirectoryIDByOrganizationExternalID :one
select scim_directories.id
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and organizations.external_id = $2
  and scim_directories.is_primary = true;

-- name: ListSCIMUsers :many
select *
from scim_users
where scim_directory_id = $1
  and id >= $2
order by id
limit $3;

-- name: ListSCIMUsersInSCIMGroup :many
select *
from scim_users
where scim_users.scim_directory_id = $1
  and scim_users.id >= $2
  and exists(select *
             from scim_user_group_memberships
             where scim_group_id = $4
               and scim_user_id = scim_users.id)
order by scim_users.id
limit $3;

-- name: GetSCIMUser :one
select scim_users.*
from scim_users
         join scim_directories
              on scim_users.scim_directory_id = scim_directories.id
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and scim_users.id = $2;

-- name: ListSCIMGroups :many
select *
from scim_groups
where scim_directory_id = $1
  and id >= $2
order by id
limit $3;

-- name: ListSCIMGroupsBySCIMUserID :many
select *
from scim_groups
where scim_groups.scim_directory_id = $1
  and scim_groups.id >= $2
  and exists(select *
             from scim_user_group_memberships
             where scim_user_group_memberships.scim_user_id = $3
               and scim_user_group_memberships.scim_group_id = scim_groups.id)
order by scim_groups.id
limit $4;

-- name: GetSCIMGroup :one
select scim_groups.*
from scim_groups
         join scim_directories
              on scim_groups.scim_directory_id = scim_directories.id
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and scim_groups.id = $2;

-- name: CreateSCIMDirectory :one
insert into scim_directories (id, organization_id, bearer_token_sha256,
                              is_primary, scim_base_url)
values ($1, $2, $3, $4, $5)
returning *;

-- name: UpdateSCIMDirectory :one
update scim_directories
set is_primary = $1
where id = $2
returning *;

-- name: UpdatePrimarySCIMDirectory :exec
update scim_directories
set is_primary = (id = $1)
where organization_id = $2;

-- name: ListSCIMDirectories :many
select *
from scim_directories
where organization_id = $1
  and id >= $2
order by id
limit $3;

-- name: GetSCIMDirectory :one
select scim_directories.*
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and scim_directories.id = $2;

-- name: ManagementGetSCIMDirectory :one
select scim_directories.*
from scim_directories
         join organizations
              on scim_directories.organization_id = organizations.id
where organizations.environment_id = $1
  and scim_directories.id = $2;

-- name: AppGetSCIMUser :one
select scim_users.*
from scim_users
         join scim_directories
              on scim_users.scim_directory_id = scim_directories.id
         join organizations
              on scim_directories.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and scim_users.id = $2;

-- name: AppGetSCIMGroup :one
select scim_groups.*
from scim_groups
         join scim_directories
              on scim_groups.scim_directory_id = scim_directories.id
         join organizations
              on scim_directories.organization_id = organizations.id
         join environments on organizations.environment_id = environments.id
where environments.app_organization_id = $1
  and scim_groups.id = $2;

-- name: UpdateSCIMDirectoryBearerToken :one
update scim_directories
set bearer_token_sha256 = $1
where id = $2
returning *;

-- name: DeleteSCIMDirectory :execrows
delete
from scim_directories
where id = $1;

-- name: DeleteSCIMUsersBySCIMDirectory :execrows
delete
from scim_users
where scim_directory_id = $1;

-- name: DeleteSCIMGroupsBySCIMDirectory :execrows
delete
from scim_groups
where scim_directory_id = $1;

-- name: DeleteSCIMUserGroupMembershipsBySCIMDirectory :execrows
delete
from scim_user_group_memberships
where scim_directory_id = $1;

-- name: DeleteSCIMRequestsBySCIMDirectory :execrows
delete
from scim_requests
where scim_directory_id = $1;
