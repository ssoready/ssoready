-- name: GetSAMLConnectionByID :one
select * from saml_connections where id = $1;
