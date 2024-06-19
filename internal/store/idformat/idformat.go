package idformat

import "github.com/ssoready/prettyuuid"

const alphabet = "0123456789abcdefghijklmnopqrstuvwxyz"

var (
	AppUser          = prettyuuid.MustNewFormat("app_user_", alphabet)
	Environment      = prettyuuid.MustNewFormat("env_", alphabet)
	APIKey           = prettyuuid.MustNewFormat("api_key_", alphabet)
	APISecretKey     = prettyuuid.MustNewFormat("ssoready_sk_", alphabet)
	Organization     = prettyuuid.MustNewFormat("org_", alphabet)
	SAMLConnection   = prettyuuid.MustNewFormat("saml_conn_", alphabet)
	SAMLFlow         = prettyuuid.MustNewFormat("saml_flow_", alphabet)
	SAMLAccessCode   = prettyuuid.MustNewFormat("saml_access_code_", alphabet)
	OAuthAccessToken = prettyuuid.MustNewFormat("ssoready_oauth_access_token_", alphabet)
)
