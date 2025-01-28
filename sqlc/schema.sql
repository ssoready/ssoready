--
-- PostgreSQL database dump
--

-- Dumped from database version 15.3 (Debian 15.3-1.pgdg120+1)
-- Dumped by pg_dump version 17.0 (Homebrew)

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
SET transaction_timeout = 0;
SET client_encoding = 'UTF8';
SET standard_conforming_strings = on;
SELECT pg_catalog.set_config('search_path', '', false);
SET check_function_bodies = false;
SET xmloption = content;
SET client_min_messages = warning;
SET row_security = off;

--
-- Name: saml_flow_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.saml_flow_status AS ENUM (
    'in_progress',
    'failed',
    'succeeded'
);


ALTER TYPE public.saml_flow_status OWNER TO postgres;

--
-- Name: scim_request_http_method; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scim_request_http_method AS ENUM (
    'get',
    'post',
    'put',
    'patch',
    'delete'
);


ALTER TYPE public.scim_request_http_method OWNER TO postgres;

--
-- Name: scim_request_http_status; Type: TYPE; Schema: public; Owner: postgres
--

CREATE TYPE public.scim_request_http_status AS ENUM (
    '200',
    '201',
    '204',
    '400',
    '401',
    '404'
);


ALTER TYPE public.scim_request_http_status OWNER TO postgres;

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: admin_access_tokens; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.admin_access_tokens (
    id uuid NOT NULL,
    organization_id uuid NOT NULL,
    one_time_token_sha256 bytea,
    access_token_sha256 bytea,
    create_time timestamp with time zone NOT NULL,
    expire_time timestamp with time zone NOT NULL,
    can_manage_saml boolean,
    can_manage_scim boolean
);


ALTER TABLE public.admin_access_tokens OWNER TO postgres;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_keys (
    id uuid NOT NULL,
    secret_value character varying NOT NULL,
    environment_id uuid NOT NULL,
    secret_value_sha256 bytea,
    has_management_api_access boolean
);


ALTER TABLE public.api_keys OWNER TO postgres;

--
-- Name: app_organizations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_organizations (
    id uuid NOT NULL,
    google_hosted_domain character varying,
    microsoft_tenant_id character varying,
    email_logins_disabled boolean,
    entitled_management_api boolean,
    entitled_custom_domains boolean
);


ALTER TABLE public.app_organizations OWNER TO postgres;

--
-- Name: app_sessions; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_sessions (
    id uuid NOT NULL,
    app_user_id uuid NOT NULL,
    create_time timestamp with time zone NOT NULL,
    expire_time timestamp with time zone NOT NULL,
    token character varying NOT NULL,
    token_sha256 bytea,
    revoked boolean
);


ALTER TABLE public.app_sessions OWNER TO postgres;

--
-- Name: app_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_users (
    id uuid NOT NULL,
    app_organization_id uuid NOT NULL,
    display_name character varying NOT NULL,
    email character varying NOT NULL
);


ALTER TABLE public.app_users OWNER TO postgres;

--
-- Name: email_verification_challenges; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.email_verification_challenges (
    id uuid NOT NULL,
    email character varying NOT NULL,
    expire_time timestamp with time zone NOT NULL,
    secret_token character varying NOT NULL,
    complete_time timestamp with time zone
);


ALTER TABLE public.email_verification_challenges OWNER TO postgres;

--
-- Name: environments; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.environments (
    id uuid NOT NULL,
    redirect_url character varying,
    app_organization_id uuid NOT NULL,
    display_name character varying,
    auth_url character varying,
    oauth_redirect_uri character varying,
    custom_auth_domain character varying,
    admin_application_name character varying,
    admin_return_url character varying,
    custom_admin_domain character varying,
    admin_url character varying,
    admin_logo_configured boolean DEFAULT false NOT NULL
);


ALTER TABLE public.environments OWNER TO postgres;

--
-- Name: onboarding_states; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.onboarding_states (
    app_organization_id uuid NOT NULL,
    dummyidp_app_id character varying NOT NULL,
    onboarding_environment_id uuid NOT NULL,
    onboarding_organization_id uuid NOT NULL,
    onboarding_saml_connection_id uuid NOT NULL
);


ALTER TABLE public.onboarding_states OWNER TO postgres;

--
-- Name: organization_domains; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.organization_domains (
    id uuid NOT NULL,
    organization_id uuid NOT NULL,
    domain character varying NOT NULL
);


ALTER TABLE public.organization_domains OWNER TO postgres;

--
-- Name: organizations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.organizations (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    external_id character varying,
    display_name character varying
);


ALTER TABLE public.organizations OWNER TO postgres;

--
-- Name: saml_connections; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.saml_connections (
    id uuid NOT NULL,
    organization_id uuid NOT NULL,
    idp_redirect_url character varying,
    idp_x509_certificate bytea,
    idp_entity_id character varying,
    sp_entity_id character varying NOT NULL,
    is_primary boolean DEFAULT false NOT NULL,
    sp_acs_url character varying NOT NULL
);


ALTER TABLE public.saml_connections OWNER TO postgres;

--
-- Name: saml_flows; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.saml_flows (
    id uuid NOT NULL,
    saml_connection_id uuid NOT NULL,
    access_code uuid,
    state character varying NOT NULL,
    create_time timestamp with time zone NOT NULL,
    expire_time timestamp with time zone NOT NULL,
    email character varying,
    subject_idp_attributes jsonb,
    update_time timestamp with time zone NOT NULL,
    auth_redirect_url character varying,
    get_redirect_time timestamp with time zone,
    initiate_request character varying,
    initiate_time timestamp with time zone,
    assertion character varying,
    app_redirect_url character varying,
    receive_assertion_time timestamp with time zone,
    redeem_time timestamp with time zone,
    redeem_response jsonb,
    error_bad_issuer character varying,
    error_bad_audience character varying,
    error_bad_subject_id character varying,
    error_email_outside_organization_domains character varying,
    status public.saml_flow_status NOT NULL,
    error_unsigned_assertion boolean DEFAULT false NOT NULL,
    access_code_sha256 bytea,
    is_oauth boolean,
    error_bad_signature_algorithm character varying,
    error_bad_digest_algorithm character varying,
    error_bad_x509_certificate bytea,
    error_saml_connection_not_configured boolean DEFAULT false NOT NULL,
    error_environment_oauth_redirect_uri_not_configured boolean DEFAULT false NOT NULL,
    assertion_id character varying,
    test_mode_idp character varying
);


ALTER TABLE public.saml_flows OWNER TO postgres;

--
-- Name: saml_oauth_clients; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.saml_oauth_clients (
    id uuid NOT NULL,
    environment_id uuid NOT NULL,
    client_secret_sha256 bytea NOT NULL
);


ALTER TABLE public.saml_oauth_clients OWNER TO postgres;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: scim_directories; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scim_directories (
    id uuid NOT NULL,
    organization_id uuid NOT NULL,
    scim_base_url character varying NOT NULL,
    bearer_token_sha256 bytea,
    is_primary boolean NOT NULL
);


ALTER TABLE public.scim_directories OWNER TO postgres;

--
-- Name: scim_groups; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scim_groups (
    id uuid NOT NULL,
    scim_directory_id uuid NOT NULL,
    display_name character varying NOT NULL,
    deleted boolean NOT NULL,
    attributes jsonb
);


ALTER TABLE public.scim_groups OWNER TO postgres;

--
-- Name: scim_requests; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scim_requests (
    id uuid NOT NULL,
    scim_directory_id uuid NOT NULL,
    "timestamp" timestamp with time zone NOT NULL,
    http_request_url character varying NOT NULL,
    http_request_method public.scim_request_http_method NOT NULL,
    http_request_body jsonb,
    http_response_status public.scim_request_http_status NOT NULL,
    http_response_body jsonb,
    error_bad_bearer_token boolean DEFAULT false NOT NULL,
    error_bad_username character varying,
    error_email_outside_organization_domains character varying
);


ALTER TABLE public.scim_requests OWNER TO postgres;

--
-- Name: scim_user_group_memberships; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scim_user_group_memberships (
    id uuid NOT NULL,
    scim_directory_id uuid NOT NULL,
    scim_user_id uuid NOT NULL,
    scim_group_id uuid NOT NULL
);


ALTER TABLE public.scim_user_group_memberships OWNER TO postgres;

--
-- Name: scim_users; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.scim_users (
    id uuid NOT NULL,
    scim_directory_id uuid NOT NULL,
    email character varying NOT NULL,
    deleted boolean NOT NULL,
    attributes jsonb
);


ALTER TABLE public.scim_users OWNER TO postgres;

--
-- Name: admin_access_tokens admin_access_tokens_access_token_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_access_tokens
    ADD CONSTRAINT admin_access_tokens_access_token_sha256_key UNIQUE (access_token_sha256);


--
-- Name: admin_access_tokens admin_access_tokens_one_time_token_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_access_tokens
    ADD CONSTRAINT admin_access_tokens_one_time_token_sha256_key UNIQUE (one_time_token_sha256);


--
-- Name: admin_access_tokens admin_access_tokens_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_access_tokens
    ADD CONSTRAINT admin_access_tokens_pkey PRIMARY KEY (id);


--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


--
-- Name: api_keys api_keys_secret_value_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_secret_value_sha256_key UNIQUE (secret_value_sha256);


--
-- Name: app_organizations app_organizations_google_hosted_domain_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_organizations
    ADD CONSTRAINT app_organizations_google_hosted_domain_key UNIQUE (google_hosted_domain);


--
-- Name: app_organizations app_organizations_google_hosted_domain_key1; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_organizations
    ADD CONSTRAINT app_organizations_google_hosted_domain_key1 UNIQUE (google_hosted_domain);


--
-- Name: app_organizations app_organizations_microsoft_tenant_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_organizations
    ADD CONSTRAINT app_organizations_microsoft_tenant_id_key UNIQUE (microsoft_tenant_id);


--
-- Name: app_organizations app_organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_organizations
    ADD CONSTRAINT app_organizations_pkey PRIMARY KEY (id);


--
-- Name: app_sessions app_sessions_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT app_sessions_pkey PRIMARY KEY (id);


--
-- Name: app_sessions app_sessions_token_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT app_sessions_token_sha256_key UNIQUE (token_sha256);


--
-- Name: app_users app_users_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_users
    ADD CONSTRAINT app_users_email_key UNIQUE (email);


--
-- Name: app_users app_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_users
    ADD CONSTRAINT app_users_pkey PRIMARY KEY (id);


--
-- Name: email_verification_challenges email_verification_challenges_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.email_verification_challenges
    ADD CONSTRAINT email_verification_challenges_pkey PRIMARY KEY (id);


--
-- Name: environments environments_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.environments
    ADD CONSTRAINT environments_pkey PRIMARY KEY (id);


--
-- Name: onboarding_states onboarding_states_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.onboarding_states
    ADD CONSTRAINT onboarding_states_pkey PRIMARY KEY (app_organization_id);


--
-- Name: organization_domains organization_domains_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organization_domains
    ADD CONSTRAINT organization_domains_pkey PRIMARY KEY (id);


--
-- Name: organizations organizations_environment_id_external_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_environment_id_external_id_key UNIQUE (environment_id, external_id);


--
-- Name: organizations organizations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_pkey PRIMARY KEY (id);


--
-- Name: saml_connections saml_connections_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_connections
    ADD CONSTRAINT saml_connections_pkey PRIMARY KEY (id);


--
-- Name: saml_flows saml_flows_access_code_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_access_code_key UNIQUE (access_code);


--
-- Name: saml_flows saml_flows_access_code_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_access_code_sha256_key UNIQUE (access_code_sha256);


--
-- Name: saml_flows saml_flows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_pkey PRIMARY KEY (id);


--
-- Name: saml_flows saml_flows_saml_connection_id_assertion_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_saml_connection_id_assertion_id_key UNIQUE (saml_connection_id, assertion_id);


--
-- Name: saml_oauth_clients saml_oauth_clients_client_secret_sha256_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_oauth_clients
    ADD CONSTRAINT saml_oauth_clients_client_secret_sha256_key UNIQUE (client_secret_sha256);


--
-- Name: saml_oauth_clients saml_oauth_clients_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_oauth_clients
    ADD CONSTRAINT saml_oauth_clients_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


--
-- Name: scim_directories scim_directories_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_directories
    ADD CONSTRAINT scim_directories_pkey PRIMARY KEY (id);


--
-- Name: scim_groups scim_groups_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_groups
    ADD CONSTRAINT scim_groups_pkey PRIMARY KEY (id);


--
-- Name: scim_requests scim_requests_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_requests
    ADD CONSTRAINT scim_requests_pkey PRIMARY KEY (id);


--
-- Name: scim_user_group_memberships scim_user_group_memberships_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_user_group_memberships
    ADD CONSTRAINT scim_user_group_memberships_pkey PRIMARY KEY (id);


--
-- Name: scim_user_group_memberships scim_user_group_memberships_scim_user_id_scim_group_id_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_user_group_memberships
    ADD CONSTRAINT scim_user_group_memberships_scim_user_id_scim_group_id_key UNIQUE (scim_user_id, scim_group_id);


--
-- Name: scim_users scim_users_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_users
    ADD CONSTRAINT scim_users_pkey PRIMARY KEY (id);


--
-- Name: scim_users scim_users_scim_directory_id_email_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_users
    ADD CONSTRAINT scim_users_scim_directory_id_email_key UNIQUE (scim_directory_id, email);


--
-- Name: admin_access_tokens admin_access_tokens_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.admin_access_tokens
    ADD CONSTRAINT admin_access_tokens_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: api_keys api_keys_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environments(id);


--
-- Name: app_sessions app_sessions_app_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT app_sessions_app_user_id_fkey FOREIGN KEY (app_user_id) REFERENCES public.app_users(id);


--
-- Name: app_users app_users_app_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_users
    ADD CONSTRAINT app_users_app_organization_id_fkey FOREIGN KEY (app_organization_id) REFERENCES public.app_organizations(id);


--
-- Name: environments environments_app_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.environments
    ADD CONSTRAINT environments_app_organization_id_fkey FOREIGN KEY (app_organization_id) REFERENCES public.app_organizations(id);


--
-- Name: onboarding_states onboarding_states_app_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.onboarding_states
    ADD CONSTRAINT onboarding_states_app_organization_id_fkey FOREIGN KEY (app_organization_id) REFERENCES public.app_organizations(id);


--
-- Name: organization_domains organization_domains_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organization_domains
    ADD CONSTRAINT organization_domains_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: organizations organizations_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.organizations
    ADD CONSTRAINT organizations_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environments(id);


--
-- Name: saml_connections saml_connections_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_connections
    ADD CONSTRAINT saml_connections_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: saml_flows saml_flows_saml_connection_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_saml_connection_id_fkey FOREIGN KEY (saml_connection_id) REFERENCES public.saml_connections(id);


--
-- Name: saml_oauth_clients saml_oauth_clients_environment_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_oauth_clients
    ADD CONSTRAINT saml_oauth_clients_environment_id_fkey FOREIGN KEY (environment_id) REFERENCES public.environments(id);


--
-- Name: scim_directories scim_directories_organization_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_directories
    ADD CONSTRAINT scim_directories_organization_id_fkey FOREIGN KEY (organization_id) REFERENCES public.organizations(id);


--
-- Name: scim_groups scim_groups_scim_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_groups
    ADD CONSTRAINT scim_groups_scim_directory_id_fkey FOREIGN KEY (scim_directory_id) REFERENCES public.scim_directories(id);


--
-- Name: scim_requests scim_requests_scim_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_requests
    ADD CONSTRAINT scim_requests_scim_directory_id_fkey FOREIGN KEY (scim_directory_id) REFERENCES public.scim_directories(id);


--
-- Name: scim_user_group_memberships scim_user_group_memberships_scim_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_user_group_memberships
    ADD CONSTRAINT scim_user_group_memberships_scim_directory_id_fkey FOREIGN KEY (scim_directory_id) REFERENCES public.scim_directories(id);


--
-- Name: scim_user_group_memberships scim_user_group_memberships_scim_group_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_user_group_memberships
    ADD CONSTRAINT scim_user_group_memberships_scim_group_id_fkey FOREIGN KEY (scim_group_id) REFERENCES public.scim_groups(id);


--
-- Name: scim_user_group_memberships scim_user_group_memberships_scim_user_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_user_group_memberships
    ADD CONSTRAINT scim_user_group_memberships_scim_user_id_fkey FOREIGN KEY (scim_user_id) REFERENCES public.scim_users(id);


--
-- Name: scim_users scim_users_scim_directory_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.scim_users
    ADD CONSTRAINT scim_users_scim_directory_id_fkey FOREIGN KEY (scim_directory_id) REFERENCES public.scim_directories(id);


--
-- PostgreSQL database dump complete
--

