![](https://i.imgur.com/OhtkhbJ.png)

<div align="center">
  <h1>SSOReady</h1>
  <a href="https://github.com/ssoready/ssoready-typescript"><img src="https://img.shields.io/npm/v/ssoready.svg?style=flat&color=ECDC68" /></a>
  <a href="https://github.com/ssoready/ssoready-python"><img src="https://img.shields.io/pypi/v/ssoready.svg?style=flat" /></a>
  <a href="https://github.com/ssoready/ssoready-go"><img src="https://img.shields.io/github/v/tag/ssoready/ssoready-go?style=flat&label=golang&color=%23007D9C" /></a>
  <a href="https://github.com/ssoready/ssoready-java"><img src="https://img.shields.io/maven-central/v/com.ssoready/ssoready-java?style=flat&label=maven&color=FD8100" /></a>
  <a href="https://github.com/ssoready/ssoready-csharp"><img src="https://img.shields.io/nuget/v/SSOReady.Client?style=flat&color=004880" /></a>
  <a href="https://github.com/ssoready/ssoready-ruby"><img src="https://img.shields.io/gem/v/ssoready?style=flat&color=EE3F2D" /></a>
  <a href="https://github.com/ssoready/ssoready-php"><img src="https://img.shields.io/packagist/v/ssoready/ssoready?style=flat&color=F28D1A" /></a>
  <a href="https://github.com/ssoready/ssoready/blob/main/LICENSE"><img src="https://img.shields.io/badge/license-MIT-blue" /></a>
  <a href="https://github.com/ssoready/ssoready/stargazers"><img src="https://img.shields.io/github/stars/ssoready/ssoready?style=flat&logo=github&color=white" /></a>
  <br />
  <br />
  <a href="https://ssoready.com/docs/saml/saml-quickstart">SAML Quickstart</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/docs/scim/scim-quickstart">SCIM Quickstart</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com">Website</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/docs">Docs</a>
  <span>&nbsp;&nbsp;•&nbsp;&nbsp;</span>
  <a href="https://ssoready.com/blog">Blog</a>
  <br />
  <hr />
</div>

## What is SSOReady?

[SSOReady](https://ssoready.com) ([YC
W24](https://www.ycombinator.com/companies/ssoready)) is an **open-source,
straightforward** way to add SAML and SCIM support to your product:

* **[SSOReady SAML](https://ssoready.com/docs/saml/saml-quickstart)**: Everything you need to add SAML ("Enterprise SSO") to your product today.
* **[SSOReady SCIM](https://ssoready.com/docs/scim/scim-quickstart)**: Everything you need to add SCIM ("Enterprise Directory Sync") to your product today.
* **[Self-serve Setup UI](https://ssoready.com/docs/idp-configuration/enabling-self-service-configuration-for-your-customers)**:
  A hosted UI your customers use to onboard themselves onto SAML and/or
  SCIM.

**With SSOReady, you're in control:**

* SSOReady can be used in *any* application, regardless of what stack you use.
  We provide language-specific SDKs as thin wrappers over a [straightforward
  HTTP
  API](https://ssoready.com/docs/api-reference/saml/redeem-saml-access-code):
  * [SSOReady-TypeScript](https://github.com/ssoready/ssoready-typescript)
  * [SSOReady-Python](https://github.com/ssoready/ssoready-python)
  * [SSOReady-Go](https://github.com/ssoready/ssoready-go)
  * [SSOReady-Java](https://github.com/ssoready/ssoready-java)
  * [SSOReady-C#](https://github.com/ssoready/ssoready-csharp)
  * [SSOReady-Ruby](https://github.com/ssoready/ssoready-ruby)
  * [SSOReady-PHP](https://github.com/ssoready/ssoready-php)
* SSOReady is just an authentication middleware layer. SSOReady doesn’t "own" your users or require any changes to your users database.
* You can use our cloud-hosted instance or [self-host yourself](https://ssoready.com/docs/self-hosting-ssoready), with the Enterprise plan giving you SLA'd support either way. 

**SSOReady can be extended with these products, available on the [Enterprise plan](https://ssoready.com/pricing):**

* [Custom Domains & Branding](https://ssoready.com/docs/ssoready-concepts/environments#custom-domains): Run
  SSOReady on a domain you control, and make your entire SAML/SCIM experience on-brand. 
* [Management API](https://ssoready.com/docs/management-api): Completely automate everything about SAML
  and SCIM programmatically at scale.
* [Enterprise Support](https://ssoready.com/pricing): SLA'd support, including for self-hosted deployments.

## Getting started

The fastest way to get started with SSOReady is to follow the quickstart for
what you want to add support for:

* [SAML Quickstart](https://ssoready.com/docs/saml/saml-quickstart)
* [SCIM Quickstart](https://ssoready.com/docs/scim/scim-quickstart)

Most folks implement SAML and SCIM in an afternoon. It only takes two lines of
code.

## Local Development

This section covers how to run SSOReady locally for development purposes.

### Prerequisites

- [Node.js](https://nodejs.org/) (see `.nvmrc` for version) - For the frontend apps (App and Admin)
- [Go](https://golang.org/) 1.21+ - For the backend services (API and Auth)
- [Docker](https://www.docker.com/) & Docker Compose - For running PostgreSQL locally
- [PostgreSQL](https://www.postgresql.org/) 15+ (via Docker)

### Initial Setup

1. **Clone the repository**

   ```bash
   git clone https://github.com/ssoready/ssoready.git
   cd ssoready
   ```

2. **Install dependencies**

   ```bash
   npm install
   cd admin && npm install && cd ..
   cd app && npm install && cd ..
   ```

3. **Set up environment variables**

   Copy `.env.example` to `.env`:

   ```bash
   cp .env.example .env
   ```

   **Important:** A working `.env` configuration with all required credentials (Google OAuth, API keys, etc.) can be found in 1Password. Copy those values into your `.env` file.

4. **Start PostgreSQL**

   ```bash
   docker compose up -d
   ```

   This starts PostgreSQL on the port specified in your `POSTGRES_PORT` env var (defaults to 5432).

5. **Run database migrations**

   ```bash
   ./bin/migrate local up
   ```

   The migrate script automatically loads your `.env` file and uses the `API_DB` connection string to connect to your database.

### Running the Services

SSOReady consists of three main services that run in parallel:

```bash
# Terminal 1: API Service (port 8081)
npm run dev:api

# Terminal 2: Auth Service (port 8080)
npm run dev:auth

# Terminal 3: App (Customer-facing app, port 8082)
npm run dev:app

# Terminal 4: Admin (Self-service setup, port 8083)
npm run dev:admin
```

Alternatively, you can run all services in parallel (requires a multiplexer like tmux):

```bash
npm run dev
```

### Service URLs

Once running, the services are available at:

- **API Service**: http://localhost:8081
- **Auth Service**: http://localhost:8080
- **App (Customer App)**: http://localhost:8082
- **Admin (Self-service)**: http://localhost:8083

### First-time Setup

1. **Create a user account**

   Visit http://localhost:8082 and sign in with Google OAuth (configured in your `.env`).

2. **Create an organization**

   Once logged in, create your first organization/environment.

3. **Test the admin self-service**

   - In the app, navigate to an organization
   - Click "Create self-service setup link"
   - Open the generated URL (e.g., http://localhost:8083/setup?one-time-token=...)
   - Configure SAML and SCIM settings

### Troubleshooting

**CSS not loading in admin app:**

If the admin app loads without styles, manually build the Tailwind CSS:

```bash
cd admin
npx tailwindcss -i ./src/index.css -o ./public/index.css
```

The `tailwind-watch` process should rebuild this automatically when running `npm run dev:admin`.

**Database connection errors:**

- Ensure Docker is running: `docker ps`
- Check the port in your connection string matches `POSTGRES_PORT`
- Restart services after changing `.env`: Stop all running services and restart them

**Google OAuth errors:**

- Verify `GOOGLE_OAUTH_CLIENT_ID` is set in `.env`
- Add authorized JavaScript origins: `http://localhost:8082`
- Add authorized redirect URIs: `http://localhost:8081/internal/google-callback`

**CORS errors:**

- Ensure `APP_API_URL=http://localhost:8081/internal/connect` (note the `/internal/connect` path)
- Regenerate config after changing `.env`: The dev scripts should do this automatically

### Development Workflow

1. Make your code changes
2. The dev servers will automatically rebuild and reload
3. Test your changes in the browser
4. Check terminal output for any errors

## How SSOReady works

This section provides a high-level overview of how SSOReady works, and how it's possible to implement SAML and SCIM in
just an afternoon. For a more thorough introduction, visit the [SAML
quickstart](https://ssoready.com/docs/saml/saml-quickstart) or the [SCIM
quickstart](https://ssoready.com/docs/scim/scim-quickstart).

### SAML in two lines of code

SAML (aka "Enterprise SSO") consists of two steps: an *initiation* step where you redirect your users to their corporate
identity provider, and a *handling* step where you log them in once you know who they are.

To initiate logins, you'll use SSOReady's [Get SAML Redirect
URL](https://ssoready.com/docs/api-reference/saml/get-saml-redirect-url) endpoint:

```typescript
// this is how you implement a "Sign in with SSO" button
const { redirectUrl } = await ssoready.saml.getSamlRedirectUrl({
  // the ID of the organization/workspace/team (whatever you call it)
  // you want to log the user into
  organizationExternalId: "..."
});

// redirect the user to `redirectUrl`...
```

You can use whatever your preferred ID is for organizations (you might call them "workspaces" or "teams") as your
`organizationExternalId`. You configure those IDs inside SSOReady, and SSOReady handles keeping track of that
organization's SAML and SCIM settings.

To handle logins, you'll use SSOReady's [Redeem SAML Access
Code](https://ssoready.com/docs/api-reference/saml/redeem-saml-access-code) endpoint:

```typescript
// this goes in your handler for POST /ssoready-callback
const { email, organizationExternalId } = await ssoready.saml.redeemSamlAccessCode({
  samlAccessCode: "saml_access_code_..."
});

// log the user in as `email` inside `organizationExternalId`...
```

You configure the URL for your `/ssoready-callback` endpoint in SSOReady.

### SCIM in one line of code

SCIM (aka "Enterprise directory sync") is basically a way for you to get a list of your customer's employees offline.

To get a customer's employees, you'll use SSOReady's [List SCIM
Users](https://ssoready.com/docs/api-reference/scim/list-scim-users) endpoint:

```typescript
const { scimUsers, nextPageToken } = await ssoready.scim.listScimUsers({
  organizationExternalId: "my_custom_external_id"
});

// create users from each scimUser
for (const { email, deleted, attributes } of scimUsers) {
  // ...
}
```

## Philosophy

We believe everyone that sells software to businesses should support enterprise
SSO. It's a huge security win for your customers.

The biggest problem with enterprise SSO is that it's way too confusing. Most
open-source SAML libraries are underdocumented messes. Every time I've tried to
implement SAML, I was constantly looking for someone to just tell me what in the
_world_ I was supposed to concretely do.

We believe that more people will implement enterprise SSO if you make it obvious
and secure by default. We are obsessed with giving every developer clarity and
security here.

Also, we believe randomly pumping up prices on security software like this is
totally unacceptable. MIT-licensing the software gives you insurance against us
ever doing that. Do whatever you want with the code. Fork us if we ever
misbehave.

## Security

If you have a security issue to report, please contact us at
security@ssoready.com.
