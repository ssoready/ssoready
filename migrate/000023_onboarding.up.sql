create table onboarding_states
(
    app_organization_id           uuid    not null primary key references app_organizations (id),

    dummyidp_app_id               varchar not null,

    -- these are deliberately not foreign keys to avoid coupling the core app to the onboarding flow
    onboarding_environment_id     uuid    not null,
    onboarding_organization_id    uuid    not null,
    onboarding_saml_connection_id uuid    not null
);
