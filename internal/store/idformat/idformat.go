package idformat

import "github.com/ssoready/prettyuuid"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

var (
	AppUser    = prettyuuid.MustNewFormat("app_user_", alphabet)
	AppSession = prettyuuid.MustNewFormat("app_session_", alphabet)

	Environment           = prettyuuid.MustNewFormat("env_", alphabet)
	APIKey                = prettyuuid.MustNewFormat("api_key_", alphabet)
	APISecretKey          = prettyuuid.MustNewFormat("ssoready_sk_", alphabet)
	SAMLOAuthClient       = prettyuuid.MustNewFormat("saml_oauth_client_", alphabet)
	SAMLOAuthClientSecret = prettyuuid.MustNewFormat("ssoready_oauth_client_secret_", alphabet)
	Organization          = prettyuuid.MustNewFormat("org_", alphabet)
	SAMLConnection        = prettyuuid.MustNewFormat("saml_conn_", alphabet)
	SAMLFlow              = prettyuuid.MustNewFormat("saml_flow_", alphabet)
	SAMLAccessCode        = prettyuuid.MustNewFormat("saml_access_code_", alphabet)

	SCIMDirectory = prettyuuid.MustNewFormat("scim_directory_", alphabet)
	SCIMUser      = prettyuuid.MustNewFormat("scim_user_", alphabet)

	AdminOneTimeToken = prettyuuid.MustNewFormat("ssoready_one_time_token_", alphabet)
)
