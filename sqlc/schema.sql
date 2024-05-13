--
-- PostgreSQL database dump
--

-- Dumped from database version 15.3 (Debian 15.3-1.pgdg120+1)
-- Dumped by pg_dump version 16.0

SET statement_timeout = 0;
SET lock_timeout = 0;
SET idle_in_transaction_session_timeout = 0;
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

SET default_tablespace = '';

SET default_table_access_method = heap;

--
-- Name: api_keys; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.api_keys (
    id uuid NOT NULL,
    secret_value character varying NOT NULL,
    environment_id uuid NOT NULL
);


ALTER TABLE public.api_keys OWNER TO postgres;

--
-- Name: app_organizations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.app_organizations (
    id uuid NOT NULL,
    google_hosted_domain character varying
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
    token character varying NOT NULL
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
    secret_token character varying NOT NULL
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
    auth_url character varying
);


ALTER TABLE public.environments OWNER TO postgres;

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
    external_id character varying
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
    status public.saml_flow_status NOT NULL
);


ALTER TABLE public.saml_flows OWNER TO postgres;

--
-- Name: schema_migrations; Type: TABLE; Schema: public; Owner: postgres
--

CREATE TABLE public.schema_migrations (
    version bigint NOT NULL,
    dirty boolean NOT NULL
);


ALTER TABLE public.schema_migrations OWNER TO postgres;

--
-- Name: api_keys api_keys_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.api_keys
    ADD CONSTRAINT api_keys_pkey PRIMARY KEY (id);


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
-- Name: app_sessions app_sessions_token_key; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.app_sessions
    ADD CONSTRAINT app_sessions_token_key UNIQUE (token);


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
-- Name: saml_flows saml_flows_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.saml_flows
    ADD CONSTRAINT saml_flows_pkey PRIMARY KEY (id);


--
-- Name: schema_migrations schema_migrations_pkey; Type: CONSTRAINT; Schema: public; Owner: postgres
--

ALTER TABLE ONLY public.schema_migrations
    ADD CONSTRAINT schema_migrations_pkey PRIMARY KEY (version);


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
-- PostgreSQL database dump complete
--

