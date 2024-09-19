alter table saml_flows add column error_bad_signature_algorithm varchar;
alter table saml_flows add column error_bad_digest_algorithm varchar;
alter table saml_flows add column error_bad_x509_certificate bytea;
