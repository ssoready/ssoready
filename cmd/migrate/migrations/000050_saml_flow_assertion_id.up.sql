alter table saml_flows add column assertion_id varchar;
alter table saml_flows add unique (saml_connection_id, assertion_id);
