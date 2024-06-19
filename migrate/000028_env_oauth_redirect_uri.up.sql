alter table environments add column oauth_redirect_uri varchar;
alter table saml_flows add column is_oauth bool;
